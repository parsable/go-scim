package core

import "strings"

// SCIM attribute contains metadata and rules for a field in SCIM resource.
type Attribute struct {
	Name            string       `json:"name"`
	Description     string       `json:"description"`
	Type            string       `json:"type"`
	SubAttributes   []*Attribute `json:"subAttributes"`
	CanonicalValues []string     `json:"canonicalValues"`
	MultiValued     bool         `json:"multiValued"`
	Required        bool         `json:"required"`
	CaseExact       bool         `json:"caseExact"`
	Mutability      string       `json:"mutability"`
	Returned        string       `json:"returned"`
	Uniqueness      string       `json:"uniqueness"`
	ReferenceTypes  []string     `json:"referenceTypes"`
	Metadata        *Metadata    `json:"-"`
}

// Set the default value in attributes.
func (attr *Attribute) setDefaults() {
	if attr == nil {
		return
	}

	if len(attr.Type) == 0 {
		attr.Type = TypeString
	}

	if len(attr.Mutability) == 0 {
		attr.Mutability = MutabilityReadWrite
	}

	if len(attr.Returned) == 0 {
		attr.Returned = ReturnedDefault
	}

	if len(attr.Uniqueness) == 0 {
		attr.Uniqueness = UniquenessNone
	}
}

// Returns true if the property that this attribute represents can be addressed
// by the name. According to SCIM spec, comparison is made against the name of the
// attribute in case insensitive fashion.
func (attr *Attribute) GoesBy(name string) bool {
	return attr != nil && strings.ToLower(name) == strings.ToLower(attr.Name)
}

// Return an attribute same to this attribute, but with multiValued set to false.
// If this attribute is not multiValued, this is returned.
func (attr *Attribute) ToSingleValued() *Attribute {
	if attr == nil {
		return nil
	}

	if !attr.MultiValued {
		return attr
	}

	return &Attribute{
		Name:            attr.Name,
		Description:     attr.Description,
		Type:            attr.Type,
		SubAttributes:   attr.SubAttributes,
		CanonicalValues: attr.CanonicalValues,
		MultiValued:     false,
		Required:        attr.Required,
		CaseExact:       attr.CaseExact,
		Mutability:      attr.Mutability,
		Returned:        attr.Returned,
		Uniqueness:      attr.Uniqueness,
		ReferenceTypes:  attr.ReferenceTypes,
		Metadata:        attr.Metadata,
	}
}

// Return an attribute same to this attribute, but with required set to false.
// If this attribute is not required, this is returned.
func (attr *Attribute) ToOptional() *Attribute {
	if attr == nil {
		return nil
	}

	if !attr.Required {
		return attr
	}

	return &Attribute{
		Name:            attr.Name,
		Description:     attr.Description,
		Type:            attr.Type,
		SubAttributes:   attr.SubAttributes,
		CanonicalValues: attr.CanonicalValues,
		MultiValued:     attr.MultiValued,
		Required:        false,
		CaseExact:       attr.CaseExact,
		Mutability:      attr.Mutability,
		Returned:        attr.Returned,
		Uniqueness:      attr.Uniqueness,
		ReferenceTypes:  attr.ReferenceTypes,
		Metadata:        attr.Metadata,
	}
}

// Make a deep copy of the attribute.
func (attr *Attribute) Copy() *Attribute {
	if attr == nil {
		return nil
	}

	var (
		subAttributes   []*Attribute = nil
		canonicalValues []string     = nil
		referenceTypes  []string     = nil
		metadata        *Metadata    = nil
	)
	{
		if len(attr.SubAttributes) > 0 {
			subAttributes = make([]*Attribute, 0)
			for _, subAttr := range attr.SubAttributes {
				subAttributes = append(subAttributes, subAttr.Copy())
			}
		}

		if len(attr.CanonicalValues) > 0 {
			canonicalValues = make([]string, 0)
			for _, cv := range attr.CanonicalValues {
				canonicalValues = append(canonicalValues, cv)
			}
		}

		if len(attr.ReferenceTypes) > 0 {
			referenceTypes = make([]string, 0)
			for _, ref := range attr.ReferenceTypes {
				referenceTypes = append(referenceTypes, ref)
			}
		}

		if attr.Metadata != nil {
			metadata = attr.Metadata.copy()
		}
	}

	return &Attribute{
		Name:            attr.Name,
		Description:     attr.Description,
		Type:            attr.Type,
		SubAttributes:   subAttributes,
		CanonicalValues: canonicalValues,
		MultiValued:     attr.MultiValued,
		Required:        attr.Required,
		CaseExact:       attr.CaseExact,
		Mutability:      attr.Mutability,
		Returned:        attr.Returned,
		Uniqueness:      attr.Uniqueness,
		ReferenceTypes:  referenceTypes,
		Metadata:        metadata,
	}
}