package configW

type mode struct {
	inAttr bool
	name   string
}

type Result struct {
	Variables  map[string]string
	Mapping    map[string]string
	Attributes map[string][]string
}
