package centraldocs

import (
	"fmt"
	"strings"

	"github.com/heshanpadmasiri/central-search/internal/catalog"
)

func normalizePage(props pageProps) (catalog.ScrapedModule, error) {
	moduleDTO := props.Module
	if strings.TrimSpace(moduleDTO.ID) == "" && moduleDTO.IsDefault {
		moduleDTO.ID = props.PackageData.DefaultModuleName
	}
	if strings.TrimSpace(moduleDTO.ID) == "" {
		return catalog.ScrapedModule{}, fmt.Errorf("malformed Central page data: module identifier is missing")
	}
	module := newModule(moduleDTO.ID, moduleDTO.IsDefault)
	module.Summary = moduleDTO.Summary
	module.Overview = moduleDTO.Description
	current := moduleDTO.ID
	for _, item := range moduleDTO.Functions {
		module.Functions = append(module.Functions, normalizeFunction(item, current, "function"))
	}
	for _, item := range moduleDTO.Resources {
		module.Resources = append(module.Resources, normalizeFunction(item, current, "resourceMethod"))
	}
	for _, item := range moduleDTO.Classes {
		module.Classes = append(module.Classes, normalizeClass(item, current))
	}
	for _, item := range moduleDTO.Clients {
		module.Clients = append(module.Clients, normalizeClient(item, current))
	}
	for _, item := range moduleDTO.Listeners {
		module.Listeners = append(module.Listeners, normalizeListener(item, current))
	}
	for _, item := range moduleDTO.ObjectTypes {
		module.Objects = append(module.Objects, normalizeObject(item, current, false))
	}
	for _, item := range moduleDTO.ServiceTypes {
		module.Services = append(module.Services, normalizeService(item, current))
	}
	for _, item := range moduleDTO.Records {
		module.Records = append(module.Records, normalizeRecord(item, current))
	}
	for _, item := range moduleDTO.Enums {
		module.Enums = append(module.Enums, normalizeEnum(item))
	}
	typeBuckets := []struct {
		kind   string
		values []typeDTO
	}{
		{"alias", moduleDTO.Types}, {"union", moduleDTO.UnionTypes}, {"alias", moduleDTO.SimpleNameReferenceTypes},
		{"tuple", moduleDTO.TupleTypes}, {"table", moduleDTO.TableTypes}, {"map", moduleDTO.MapTypes},
		{"intersection", moduleDTO.IntersectionTypes}, {"typedesc", moduleDTO.TypeDescriptorTypes}, {"function", moduleDTO.FunctionTypes},
		{"stream", moduleDTO.StreamTypes}, {"array", moduleDTO.ArrayTypes}, {"anydata", moduleDTO.AnyDataTypes},
		{"any", moduleDTO.AnyTypes}, {"string", moduleDTO.StringTypes}, {"int", moduleDTO.IntegerTypes},
		{"decimal", moduleDTO.DecimalTypes}, {"xml", moduleDTO.XMLTypes}, {"boolean", moduleDTO.BooleanTypes},
	}
	for _, bucket := range typeBuckets {
		for _, item := range bucket.values {
			module.Types = append(module.Types, normalizeType(item, current, bucket.kind))
		}
	}
	for _, item := range moduleDTO.Errors {
		module.Errors = append(module.Errors, normalizeError(item, current))
	}
	for _, item := range moduleDTO.Constants {
		module.Constants = append(module.Constants, normalizeConstant(item, current))
	}
	for _, item := range moduleDTO.Variables {
		module.Variables = append(module.Variables, normalizeVariable(item, current, false))
	}
	for _, item := range moduleDTO.Configurables {
		module.Configurables = append(module.Configurables, normalizeVariable(item, current, true))
	}
	for _, item := range moduleDTO.Annotations {
		module.Annotations = append(module.Annotations, normalizeAnnotation(item, current))
	}
	related := make([]catalog.ModuleReference, 0, len(moduleDTO.RelatedModules))
	for _, item := range moduleDTO.RelatedModules {
		if item.ID != "" {
			related = append(related, catalog.ModuleReference{Name: item.ID, IsDefault: item.IsDefault})
		}
	}
	return catalog.ScrapedModule{PackageSummary: props.PackageData.Summary, PackageOverview: props.PackageData.Readme, Module: module, RelatedModules: related}, nil
}

func newModule(name string, defaultModule bool) catalog.Module {
	return catalog.Module{Name: name, IsDefault: defaultModule,
		Functions: []catalog.Function{}, Resources: []catalog.Function{}, Classes: []catalog.Class{}, Clients: []catalog.Client{}, Listeners: []catalog.Listener{}, Objects: []catalog.Object{}, Services: []catalog.ServiceDeclaration{}, Records: []catalog.Record{}, Enums: []catalog.Enum{}, Types: []catalog.TypeDefinition{}, Errors: []catalog.ErrorType{}, Constants: []catalog.Constant{}, Variables: []catalog.Variable{}, Configurables: []catalog.Variable{}, Annotations: []catalog.Annotation{}}
}

func documentation(base baseDTO) string {
	parts := make([]string, 0, len(base.DescriptionSections)+1)
	if strings.TrimSpace(base.Description) != "" {
		parts = append(parts, strings.TrimRight(base.Description, "\n"))
	}
	for _, section := range base.DescriptionSections {
		if strings.TrimSpace(section) != "" {
			parts = append(parts, strings.TrimRight(section, "\n"))
		}
	}
	return strings.Join(parts, "\n\n")
}

func symbol(base baseDTO, kind, signature string) catalog.Symbol {
	return catalog.Symbol{Kind: kind, Name: base.Name, Signature: signature, Documentation: documentation(base), Deprecated: base.Deprecated, ReadOnly: base.ReadOnly}
}

func normalizeFunction(item callableDTO, module, kind string) catalog.Function {
	result := catalog.Function{Symbol: symbol(item.baseDTO, kind, callableSignature(item, module, kind)), Arguments: []catalog.Argument{}, Returns: []catalog.ReturnValue{}, Accessor: item.Accessor, ResourcePath: item.ResourcePath, Isolated: item.Isolated, External: item.External, Remote: item.Remote, Resource: item.Resource}
	for _, parameter := range item.Parameters {
		result.Arguments = append(result.Arguments, catalog.Argument{Name: parameter.Name, Type: typeSignature(parameter.Type, module), Signature: parameterSignature(parameter, module), Documentation: parameter.Description, DefaultValue: parameter.DefaultValue, Optional: parameter.Type.Optional, Rest: parameter.Type.Rest})
	}
	for _, returned := range item.Returns {
		result.Returns = append(result.Returns, catalog.ReturnValue{Name: returned.Name, Type: typeSignature(returned.Type, module), Documentation: returned.Description})
	}
	return result
}

func normalizeField(item fieldDTO, module, kind string) catalog.Field {
	t := typeDTO{}
	if item.Type != nil {
		t = *item.Type
	}
	typeName := typeSignature(t, module)
	signature := strings.TrimSpace(typeName + " " + item.Name)
	if t.Optional {
		signature += "?"
	}
	if item.DefaultValue != "" {
		signature += " = " + item.DefaultValue
	}
	return catalog.Field{Symbol: symbol(item.baseDTO, kind, signature), Type: typeName, Optional: t.Optional, DefaultValue: item.DefaultValue}
}

func ordinaryMethods(methods, explicit []callableDTO) []callableDTO {
	if explicit != nil {
		return explicit
	}
	result := make([]callableDTO, 0)
	for _, method := range methods {
		if method.Name != "init" && !method.Remote && !method.Resource {
			result = append(result, method)
		}
	}
	return result
}
func findInit(methods []callableDTO, explicit *callableDTO) *callableDTO {
	if explicit != nil && explicit.Name != "" {
		return explicit
	}
	for i := range methods {
		if methods[i].Name == "init" {
			return &methods[i]
		}
	}
	return nil
}
func normalizeMethods(items []callableDTO, module, kind string) []catalog.Function {
	result := make([]catalog.Function, 0, len(items))
	for _, item := range items {
		result = append(result, normalizeFunction(item, module, kind))
	}
	return result
}
func normalizeFields(items []fieldDTO, module string) []catalog.Field {
	result := make([]catalog.Field, 0, len(items))
	for _, item := range items {
		if item.InclusionType == nil {
			result = append(result, normalizeField(item, module, "field"))
		}
	}
	return result
}

func normalizeClass(item classDTO, module string) catalog.Class {
	methods := ordinaryMethods(item.Methods, item.OtherMethods)
	result := catalog.Class{Symbol: symbol(item.baseDTO, "class", namedSignature("public class", item.Name, "")), Fields: normalizeFields(item.Fields, module), Methods: normalizeMethods(methods, module, "method"), Isolated: item.Isolated}
	if init := findInit(item.Methods, item.Init); init != nil {
		value := normalizeFunction(*init, module, "method")
		result.Init = &value
	}
	return result
}
func normalizeClient(item clientDTO, module string) catalog.Client {
	result := catalog.Client{Symbol: symbol(item.baseDTO, "client", namedSignature("public client class", item.Name, "")), Fields: normalizeFields(item.Fields, module), Methods: normalizeMethods(ordinaryMethods(item.Methods, item.OtherMethods), module, "method"), RemoteMethods: normalizeMethods(item.RemoteMethods, module, "remoteMethod"), ResourceMethods: normalizeMethods(item.ResourceMethods, module, "resourceMethod"), Isolated: item.Isolated}
	if init := findInit(item.Methods, item.Init); init != nil {
		value := normalizeFunction(*init, module, "method")
		result.Init = &value
	}
	return result
}
func normalizeListener(item listenerDTO, module string) catalog.Listener {
	result := catalog.Listener{Symbol: symbol(item.baseDTO, "listener", namedSignature("public class", item.Name, "")), Fields: normalizeFields(item.Fields, module), Methods: normalizeMethods(ordinaryMethods(item.Methods, item.OtherMethods), module, "method"), LifecycleMethods: normalizeMethods(item.LifecycleMethods, module, "method"), Isolated: item.Isolated}
	if init := findInit(item.Methods, item.Init); init != nil {
		value := normalizeFunction(*init, module, "method")
		result.Init = &value
	}
	return result
}
func normalizeObject(item objectDTO, module string, _ bool) catalog.Object {
	includes := inclusions(item.Fields, module)
	return catalog.Object{Symbol: symbol(item.baseDTO, "object", namedSignature("public type", item.Name, "object {}")), Fields: normalizeFields(item.Fields, module), Methods: normalizeMethods(item.Methods, module, "method"), Includes: includes, Distinct: item.Distinct}
}
func normalizeService(item objectDTO, module string) catalog.ServiceDeclaration {
	return catalog.ServiceDeclaration{Symbol: symbol(item.baseDTO, "service", namedSignature("public type", item.Name, "service object {}")), Fields: normalizeFields(item.Fields, module), Methods: normalizeMethods(item.Methods, module, "method"), Includes: inclusions(item.Fields, module), Distinct: item.Distinct}
}
func inclusions(fields []fieldDTO, module string) []string {
	result := []string{}
	for _, field := range fields {
		if field.InclusionType != nil {
			result = append(result, typeSignature(*field.InclusionType, module))
		}
	}
	return result
}
func normalizeRecord(item recordDTO, module string) catalog.Record {
	fields := []catalog.RecordField{}
	for _, field := range item.Fields {
		if field.InclusionType != nil {
			continue
		}
		f := normalizeField(field, module, "recordField")
		fields = append(fields, catalog.RecordField{Symbol: f.Symbol, Type: f.Type, Optional: f.Optional, DefaultValue: f.DefaultValue})
	}
	includes := inclusions(item.Fields, module)
	body := recordSignature(item, module)
	return catalog.Record{Symbol: symbol(item.baseDTO, "record", body), Closed: item.Closed, Includes: includes, Fields: fields}
}
func recordSignature(item recordDTO, module string) string {
	parts := []string{}
	for _, inc := range inclusions(item.Fields, module) {
		parts = append(parts, "*"+inc+";")
	}
	for _, f := range item.Fields {
		if f.InclusionType == nil {
			parts = append(parts, normalizeField(f, module, "recordField").Signature+";")
		}
	}
	delims := "record { "
	end := " }"
	if item.Closed {
		delims = "record {| "
		end = " |}"
	}
	return "public type " + item.Name + " " + delims + strings.Join(parts, " ") + end
}
func normalizeEnum(item enumDTO) catalog.Enum {
	members := make([]catalog.EnumMember, 0, len(item.Members))
	for _, m := range item.Members {
		sig := m.Name
		if m.Value != "" {
			sig += " = " + m.Value
		}
		members = append(members, catalog.EnumMember{Symbol: symbol(m.baseDTO, "enumMember", sig), Value: m.Value})
	}
	names := make([]string, len(members))
	for i := range members {
		names[i] = members[i].Name
	}
	return catalog.Enum{Symbol: symbol(item.baseDTO, "enum", "public enum "+item.Name+" { "+strings.Join(names, ", ")+" }"), Members: members}
}
func normalizeType(item typeDTO, module, kind string) catalog.TypeDefinition {
	copy := item
	copy.Name = ""
	definition := typeSignature(copy, module)
	members := make([]string, len(item.MemberTypes))
	for i := range item.MemberTypes {
		members[i] = typeSignature(item.MemberTypes[i], module)
	}
	return catalog.TypeDefinition{Symbol: symbol(item.baseDTO, "type", typeDeclarationSignature(item, module, kind)), TypeKind: kind, Definition: definition, Members: members}
}
func normalizeError(item errorDTO, module string) catalog.ErrorType {
	detail := typeSignature(item.DetailType, module)
	prefix := "public type "
	if item.Distinct {
		prefix += "distinct "
	}
	return catalog.ErrorType{Symbol: symbol(item.baseDTO, "error", prefix+item.Name+" error<"+detail+">"), DetailType: detail, Distinct: item.Distinct}
}
func normalizeConstant(item constantDTO, module string) catalog.Constant {
	t := typeSignature(item.Type, module)
	return catalog.Constant{Symbol: symbol(item.baseDTO, "constant", "public const "+t+" "+item.Name+" = "+item.Value), Type: t, Value: item.Value}
}
func normalizeVariable(item variableDTO, module string, configurable bool) catalog.Variable {
	t := typeSignature(item.Type, module)
	kind := "variable"
	prefix := "public "
	if configurable {
		kind = "configurable"
		prefix = "configurable "
	}
	sig := prefix + t + " " + item.Name
	if item.DefaultValue != "" {
		sig += " = " + item.DefaultValue
	}
	return catalog.Variable{Symbol: symbol(item.baseDTO, kind, sig), Type: t, DefaultValue: item.DefaultValue}
}
func normalizeAnnotation(item annotationDTO, module string) catalog.Annotation {
	points := []string{}
	for _, point := range strings.Split(item.AttachmentPoints, ",") {
		if value := strings.TrimSpace(point); value != "" {
			points = append(points, value)
		}
	}
	t := typeSignature(item.Type, module)
	sig := "public annotation " + t + " " + item.Name
	if len(points) > 0 {
		sig += " on " + strings.Join(points, ", ")
	}
	return catalog.Annotation{Symbol: symbol(item.baseDTO, "annotation", sig), Type: t, AttachmentPoints: points}
}
