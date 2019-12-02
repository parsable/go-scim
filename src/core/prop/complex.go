package prop

import (
	"encoding/binary"
	"github.com/imulab/go-scim/src/core"
	"github.com/imulab/go-scim/src/core/errors"
	"hash/fnv"
	"strings"
)

// Create a new unassigned complex property. The method will panic if
// given attribute is not singular complex type.
func NewComplex(attr *core.Attribute) core.Property {
	if !attr.SingleValued() || attr.Type() != core.TypeComplex {
		panic("invalid attribute for complex property")
	}

	var (
		subProps  = make([]core.Property, 0, attr.CountSubAttributes())
		nameIndex = make(map[string]int)
	)
	{
		attr.ForEachSubAttribute(func(subAttribute *core.Attribute) {
			if subAttribute.MultiValued() {
				subProps = append(subProps, NewMulti(subAttribute))
			} else {
				switch subAttribute.Type() {
				case core.TypeString:
					subProps = append(subProps, NewString(subAttribute))
				case core.TypeInteger:
					subProps = append(subProps, NewInteger(subAttribute))
				case core.TypeDecimal:
					subProps = append(subProps, NewDecimal(subAttribute))
				case core.TypeBoolean:
					subProps = append(subProps, NewBoolean(subAttribute))
				case core.TypeDateTime:
					subProps = append(subProps, NewDateTime(subAttribute))
				case core.TypeReference:
					subProps = append(subProps, NewReference(subAttribute))
				case core.TypeBinary:
					subProps = append(subProps, NewBinary(subAttribute))
				case core.TypeComplex:
					subProps = append(subProps, NewComplex(subAttribute))
				default:
					panic("invalid type")
				}
			}
			nameIndex[strings.ToLower(subAttribute.Name())] = len(subProps) - 1
		})
	}

	return &complexProperty{
		attr:      attr,
		subProps:  subProps,
		nameIndex: nameIndex,
	}
}

// Create a new complex property with given value. The method will panic if
// given attribute is not singular complex type. The property will be
// marked dirty at the start unless value is empty
func NewComplexOf(attr *core.Attribute, value interface{}) core.Property {
	p := NewComplex(attr)
	if err := p.Add(value); err != nil {
		panic(err)
	}
	return p
}

var (
	_ core.Property  = (*complexProperty)(nil)
	_ core.Container = (*complexProperty)(nil)
)

type complexProperty struct {
	attr      *core.Attribute
	subProps  []core.Property // array of sub properties to maintain determinate iteration order
	nameIndex map[string]int  // attribute's name (to lower case) to index in subProps to allow fast access
	hash      uint64
}

func (p *complexProperty) Attribute() *core.Attribute {
	return p.attr
}

// Caution: slow operation
func (p *complexProperty) Raw() interface{} {
	values := make(map[string]interface{})
	_ = p.ForEachChild(func(_ int, child core.Property) error {
		values[child.Attribute().Name()] = child.Raw()
		return nil
	})
	return values
}

func (p *complexProperty) IsUnassigned() bool {
	for _, prop := range p.subProps {
		if !prop.IsUnassigned() {
			return false
		}
	}
	return true
}

func (p *complexProperty) Matches(another core.Property) bool {
	if !p.attr.Equals(another.Attribute()) {
		return false
	}

	// Usually this won't happen, but still check it to be sure.
	if p.CountChildren() != another.(core.Container).CountChildren() {
		return false
	}

	return p.Hash() == another.Hash()
}

func (p *complexProperty) Hash() uint64 {
	if p.hash == 0 {
		p.computeHash()
	}
	return p.hash
}

func (p *complexProperty) EqualsTo(value interface{}) (bool, error) {
	return false, p.errIncompatibleOp()
}

func (p *complexProperty) StartsWith(value string) (bool, error) {
	return false, p.errIncompatibleOp()
}

func (p *complexProperty) EndsWith(value string) (bool, error) {
	return false, p.errIncompatibleOp()
}

func (p *complexProperty) Contains(value string) (bool, error) {
	return false, p.errIncompatibleOp()
}

func (p *complexProperty) GreaterThan(value interface{}) (bool, error) {
	return false, p.errIncompatibleOp()
}

func (p *complexProperty) LessThan(value interface{}) (bool, error) {
	return false, p.errIncompatibleOp()
}

func (p *complexProperty) Present() bool {
	for _, prop := range p.subProps {
		if prop.Present() {
			return true
		}
	}
	return false
}

func (p *complexProperty) Add(value interface{}) error {
	if value == nil {
		return nil
	}

	if m, ok := value.(map[string]interface{}); !ok {
		return p.errIncompatibleValue(value)
	} else {
		for k, v := range m {
			i, ok := p.nameIndex[strings.ToLower(k)]
			if !ok {
				continue
			}
			if err := p.subProps[i].Add(v); err != nil {
				return err
			}
		}
		p.computeHash()
		return nil
	}
}

func (p *complexProperty) Replace(value interface{}) (err error) {
	if value == nil {
		return nil
	}

	defer func() {
		if r := recover(); r != nil {
			err = p.errIncompatibleValue(value)
		}
	}()

	err = p.Delete()
	if err != nil {
		return
	}

	err = p.Add(value)
	if err != nil {
		return
	}

	return
}

func (p *complexProperty) Delete() error {
	for _, subProp := range p.subProps {
		if err := subProp.Delete(); err != nil {
			return err
		}
	}
	p.computeHash()
	return nil
}

func (p *complexProperty) Touched() bool {
	for _, subProp := range p.subProps {
		if subProp.Touched() {
			return true
		}
	}
	return false
}

func (p *complexProperty) CountChildren() int {
	return len(p.subProps)
}

func (p *complexProperty) ForEachChild(callback func(index int, child core.Property) error) error {
	for i, prop := range p.subProps {
		if err := callback(i, prop); err != nil {
			return err
		}
	}
	return nil
}

func (p *complexProperty) NewChild() int {
	return -1
}

func (p *complexProperty) ChildAtIndex(index interface{}) core.Property {
	name, ok := index.(string)
	if !ok {
		return nil
	}

	i, ok := p.nameIndex[strings.ToLower(name)]
	if !ok {
		return nil
	}

	return p.subProps[i]
}

func (p *complexProperty) Compact() {}

func (p *complexProperty) computeHash() {
	h := fnv.New64a()

	hasIdentity := p.attr.HasIdentitySubAttributes()
	err := p.ForEachChild(func(_ int, child core.Property) error {
		// Include fields in the computation if
		// - no sub attributes are marked as identity
		// - this sub attribute is marked identity
		if hasIdentity && !child.Attribute().IsIdentity() {
			return nil
		}

		_, err := h.Write([]byte(child.Attribute().Name()))
		if err != nil {
			return err
		}

		// Skip the value hash if it is unassigned
		if !child.IsUnassigned() {
			b := make([]byte, 8)
			binary.LittleEndian.PutUint64(b, child.Hash())
			_, err := h.Write(b)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		panic("error computing hash")
	}

	p.hash = h.Sum64()
}

func (p *complexProperty) errIncompatibleValue(value interface{}) error {
	return errors.InvalidValue("value of type %T is incompatible with attribute '%s'", value, p.attr.Path())
}

func (p *complexProperty) errIncompatibleOp() error {
	return errors.Internal("incompatible operation")
}
