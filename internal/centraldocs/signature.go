package centraldocs

import (
	"fmt"
	"strings"
)

func typeSignature(t typeDTO, currentModule string) string {
	var value string
	switch {
	case t.AnonymousUnion:
		value = joinTypes(t.MemberTypes, currentModule, "|")
	case t.Intersection:
		value = joinTypes(t.MemberTypes, currentModule, " & ")
	case t.Tuple:
		value = "[" + joinTypes(t.MemberTypes, currentModule, ", ") + "]"
	case t.TypeDesc:
		value = constrained("typedesc", t, currentModule)
	case t.Array:
		if t.ElementType != nil {
			value = typeSignature(*t.ElementType, currentModule)
		} else {
			value = t.Name
		}
	case t.Category == "map" || t.Name == "map":
		value = constrained("map", t, currentModule)
	case t.Category == "table" || t.Name == "table":
		value = constrained("table", t, currentModule)
	case t.Category == "stream" || t.Name == "stream":
		value = constrained("stream", t, currentModule)
	case t.Category == "function" || len(t.ParameterTypes) > 0 || t.ReturnType != nil:
		parts := make([]string, len(t.ParameterTypes))
		for i := range t.ParameterTypes {
			parts[i] = typeSignature(t.ParameterTypes[i], currentModule)
		}
		value = "function (" + strings.Join(parts, ", ") + ")"
		if t.ReturnType != nil {
			value += " returns " + typeSignature(*t.ReturnType, currentModule)
		}
	default:
		value = strings.TrimSpace(t.Name)
		if value == "" && t.ElementType != nil {
			value = typeSignature(*t.ElementType, currentModule)
		}
		if value == "" && len(t.MemberTypes) > 0 {
			value = joinTypes(t.MemberTypes, currentModule, "|")
		}
		if t.ModuleName != "" && t.ModuleName != currentModule && !strings.Contains(value, ":") {
			alias := t.ModuleName
			if index := strings.LastIndex(alias, "."); index >= 0 {
				alias = alias[index+1:]
			}
			value = alias + ":" + value
		}
	}
	if value == "" {
		value = "any"
	}
	if t.Parenthesized {
		value = "(" + value + ")"
	}
	dimensions := t.ArrayDimensions
	if t.Array && dimensions == 0 {
		dimensions = 1
	}
	value += strings.Repeat("[]", dimensions)
	if t.Nullable && !strings.HasSuffix(value, "?") {
		value += "?"
	}
	return value
}

func constrained(name string, t typeDTO, module string) string {
	if t.ElementType != nil {
		return name + "<" + typeSignature(*t.ElementType, module) + ">"
	}
	if len(t.MemberTypes) > 0 {
		return name + "<" + joinTypes(t.MemberTypes, module, ", ") + ">"
	}
	return name
}

func joinTypes(types []typeDTO, module, separator string) string {
	parts := make([]string, len(types))
	for i := range types {
		parts[i] = typeSignature(types[i], module)
	}
	return strings.Join(parts, separator)
}

func parameterSignature(parameter parameterDTO, module string) string {
	typeName := typeSignature(parameter.Type, module)
	if parameter.Type.Rest {
		typeName += "..."
	}
	name := parameter.Name
	if parameter.Type.Optional && name != "" {
		name += "?"
	}
	value := strings.TrimSpace(strings.TrimSpace(typeName) + " " + name)
	if parameter.DefaultValue != "" {
		value += " = " + parameter.DefaultValue
	}
	return value
}

func callableSignature(function callableDTO, module, kind string) string {
	qualifiers := make([]string, 0, 3)
	if function.Isolated {
		qualifiers = append(qualifiers, "isolated")
	}
	if function.Remote {
		qualifiers = append(qualifiers, "remote")
	}
	if function.Resource {
		qualifiers = append(qualifiers, "resource")
	}
	qualifiers = append(qualifiers, "function")
	name := function.Name
	if function.Resource {
		name = strings.TrimSpace(function.Accessor + " " + function.ResourcePath)
	}
	args := make([]string, len(function.Parameters))
	for i := range function.Parameters {
		args[i] = parameterSignature(function.Parameters[i], module)
	}
	result := strings.Join(qualifiers, " ") + " " + name + "(" + strings.Join(args, ", ") + ")"
	if len(function.Returns) > 0 {
		returns := make([]string, len(function.Returns))
		for i := range function.Returns {
			returns[i] = typeSignature(function.Returns[i].Type, module)
		}
		returnType := strings.Join(returns, ", ")
		if len(returns) > 1 {
			returnType = "[" + returnType + "]"
		}
		result += " returns " + returnType
	}
	if function.External {
		result += " = external"
	}
	return strings.TrimSpace(result)
}

func namedSignature(qualifier, name, body string) string {
	result := strings.TrimSpace(qualifier + " " + name)
	if body != "" {
		result += " " + body
	}
	return result
}

func typeDeclarationSignature(t typeDTO, module, kind string) string {
	copy := t
	copy.Name = ""
	definition := typeSignature(copy, module)
	return fmt.Sprintf("public type %s %s", t.Name, definition)
}
