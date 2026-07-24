package centraldocs

// These DTOs intentionally model only the subset of Central's Next.js data
// consumed by the normalizer. encoding/json ignores website additions.
type nextData struct {
	Props struct {
		PageProps struct {
			PropsData struct {
				Props pageProps `json:"props"`
			} `json:"propsData"`
		} `json:"pageProps"`
	} `json:"props"`
}

type pageProps struct {
	Module      pageModule  `json:"module"`
	PackageData packageData `json:"packageData"`
}

type packageData struct {
	Organization      string `json:"organization"`
	Name              string `json:"name"`
	Version           string `json:"version"`
	Summary           string `json:"summary"`
	Readme            string `json:"readme"`
	DefaultModuleName string `json:"defaultModuleName"`
}

type pageModule struct {
	RelatedModules           []relatedModule `json:"relatedModules"`
	Records                  []recordDTO     `json:"records"`
	Classes                  []classDTO      `json:"classes"`
	ObjectTypes              []objectDTO     `json:"objectTypes"`
	ServiceTypes             []objectDTO     `json:"serviceTypes"`
	Clients                  []clientDTO     `json:"clients"`
	Listeners                []listenerDTO   `json:"listeners"`
	Functions                []callableDTO   `json:"functions"`
	Resources                []callableDTO   `json:"resources"`
	Constants                []constantDTO   `json:"constants"`
	Annotations              []annotationDTO `json:"annotations"`
	Errors                   []errorDTO      `json:"errors"`
	Types                    []typeDTO       `json:"types"`
	UnionTypes               []typeDTO       `json:"unionTypes"`
	SimpleNameReferenceTypes []typeDTO       `json:"simpleNameReferenceTypes"`
	TupleTypes               []typeDTO       `json:"tupleTypes"`
	TableTypes               []typeDTO       `json:"tableTypes"`
	MapTypes                 []typeDTO       `json:"mapTypes"`
	IntersectionTypes        []typeDTO       `json:"intersectionTypes"`
	TypeDescriptorTypes      []typeDTO       `json:"typeDescriptorTypes"`
	FunctionTypes            []typeDTO       `json:"functionTypes"`
	StreamTypes              []typeDTO       `json:"streamTypes"`
	ArrayTypes               []typeDTO       `json:"arrayTypes"`
	AnyDataTypes             []typeDTO       `json:"anyDataTypes"`
	AnyTypes                 []typeDTO       `json:"anyTypes"`
	StringTypes              []typeDTO       `json:"stringTypes"`
	IntegerTypes             []typeDTO       `json:"integerTypes"`
	DecimalTypes             []typeDTO       `json:"decimalTypes"`
	XMLTypes                 []typeDTO       `json:"xmlTypes"`
	BooleanTypes             []typeDTO       `json:"booleanTypes"`
	Enums                    []enumDTO       `json:"enums"`
	Variables                []variableDTO   `json:"variables"`
	Configurables            []variableDTO   `json:"configurables"`
	ID                       string          `json:"id"`
	Summary                  string          `json:"summary"`
	Description              string          `json:"description"`
	OrgName                  string          `json:"orgName"`
	Version                  string          `json:"version"`
	IsDefault                bool            `json:"isDefaultModule"`
}

type relatedModule struct {
	ID        string `json:"id"`
	IsDefault bool   `json:"isDefaultModule"`
}
type baseDTO struct {
	Name                string   `json:"name"`
	Description         string   `json:"description"`
	DescriptionSections []string `json:"descriptionSections"`
	Deprecated          bool     `json:"isDeprecated"`
	ReadOnly            bool     `json:"isReadOnly"`
}
type fieldDTO struct {
	baseDTO
	Type          *typeDTO `json:"type"`
	InclusionType *typeDTO `json:"inclusionType"`
	DefaultValue  string   `json:"defaultValue"`
}
type parameterDTO struct {
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Type         typeDTO `json:"type"`
	DefaultValue string  `json:"defaultValue"`
	Deprecated   bool    `json:"isDeprecated"`
	ReadOnly     bool    `json:"isReadOnly"`
}
type returnDTO struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Type        typeDTO `json:"type"`
}
type callableDTO struct {
	baseDTO
	Accessor     string         `json:"accessor"`
	ResourcePath string         `json:"resourcePath"`
	Isolated     bool           `json:"isIsolated"`
	Remote       bool           `json:"isRemote"`
	Resource     bool           `json:"isResource"`
	External     bool           `json:"isExtern"`
	Parameters   []parameterDTO `json:"parameters"`
	Returns      []returnDTO    `json:"returnParameters"`
}
type recordDTO struct {
	baseDTO
	Fields []fieldDTO `json:"fields"`
	Closed bool       `json:"isClosed"`
}
type classDTO struct {
	baseDTO
	Fields       []fieldDTO    `json:"fields"`
	Methods      []callableDTO `json:"methods"`
	Init         *callableDTO  `json:"initMethod"`
	OtherMethods []callableDTO `json:"otherMethods"`
	Isolated     bool          `json:"isIsolated"`
}
type clientDTO struct {
	baseDTO
	Fields          []fieldDTO    `json:"fields"`
	Methods         []callableDTO `json:"methods"`
	Init            *callableDTO  `json:"initMethod"`
	OtherMethods    []callableDTO `json:"otherMethods"`
	RemoteMethods   []callableDTO `json:"remoteMethods"`
	ResourceMethods []callableDTO `json:"resourceMethods"`
	Isolated        bool          `json:"isIsolated"`
}
type listenerDTO struct {
	baseDTO
	Fields           []fieldDTO    `json:"fields"`
	Methods          []callableDTO `json:"methods"`
	Init             *callableDTO  `json:"initMethod"`
	OtherMethods     []callableDTO `json:"otherMethods"`
	LifecycleMethods []callableDTO `json:"lifeCycleMethods"`
	Isolated         bool          `json:"isIsolated"`
}
type objectDTO struct {
	baseDTO
	Fields   []fieldDTO    `json:"fields"`
	Methods  []callableDTO `json:"methods"`
	Distinct bool          `json:"isDistinct"`
}
type enumDTO struct {
	baseDTO
	Members []enumMemberDTO `json:"members"`
}
type enumMemberDTO struct {
	baseDTO
	Value string `json:"value"`
}
type errorDTO struct {
	baseDTO
	DetailType typeDTO `json:"detailType"`
	Distinct   bool    `json:"isDistinct"`
}
type constantDTO struct {
	baseDTO
	Type  typeDTO `json:"type"`
	Value string  `json:"value"`
}
type variableDTO struct {
	baseDTO
	Type         typeDTO `json:"type"`
	DefaultValue string  `json:"defaultValue"`
}
type annotationDTO struct {
	baseDTO
	Type             typeDTO `json:"type"`
	AttachmentPoints string  `json:"attachmentPoints"`
}

type typeDTO struct {
	baseDTO
	OrgName         string    `json:"orgName"`
	ModuleName      string    `json:"moduleName"`
	Category        string    `json:"category"`
	AnonymousUnion  bool      `json:"isAnonymousUnionType"`
	Inclusion       bool      `json:"isInclusion"`
	Array           bool      `json:"isArrayType"`
	Nullable        bool      `json:"isNullable"`
	Optional        bool      `json:"isOptional"`
	Tuple           bool      `json:"isTuple"`
	Intersection    bool      `json:"isIntersectionType"`
	Parenthesized   bool      `json:"isParenthesisedType"`
	TypeDesc        bool      `json:"isTypeDesc"`
	Rest            bool      `json:"isRestParam"`
	ArrayDimensions int       `json:"arrayDimensions"`
	MemberTypes     []typeDTO `json:"memberTypes"`
	ElementType     *typeDTO  `json:"elementType"`
	ParameterTypes  []typeDTO `json:"parameterTypes"`
	ReturnType      *typeDTO  `json:"returnType"`
}
