package main

type AttributeType int

const (
	AttributeStrength AttributeType = iota
	AttributeStamina
	AttributeAgility
)

func (a AttributeType) String() string {
	switch a {
	case AttributeStrength:
		return "Strength"
	case AttributeStamina:
		return "Stamina"
	case AttributeAgility:
		return "Agility"
	}
	return "Unknown"
}

type Attribute struct {
	Type AttributeType
	Val  int
}

type AttributeContainer struct {
	Attributes []*Attribute
}

func (p *AttributeContainer) Get(a AttributeType) int {
	for _, v := range p.Attributes {
		if a == v.Type {
			return v.Val
		}
	}

	return 0
}

func (p *AttributeContainer) Set(a AttributeType, val int) {
	for _, v := range p.Attributes {
		if v.Type == a {
			v.Val = val
			return
		}
	}

	p.Attributes = append(p.Attributes, &Attribute{
		Type: a,
		Val:  val,
	})
}

func (p *AttributeContainer) Modify(a AttributeType, modifier int) {
	p.Set(a, p.Get(a)+modifier)
}
