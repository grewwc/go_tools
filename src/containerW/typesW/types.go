package typesW

type Comparable interface {
	Compare(interface{}) int
}

type IntComparable int

func (i IntComparable) Compare(other interface{}) int {
	return int(i - other.(IntComparable))
}
