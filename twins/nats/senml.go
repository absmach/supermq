package nats

// Record represents one senML record.
type Record struct {
	XMLName     *bool    `json:"-" xml:"senml" cbor:"-"`
	Link        string   `json:"l,omitempty"  xml:"l,attr,omitempty" cbor:"-"`
	BaseName    string   `json:"bn,omitempty" xml:"bn,attr,omitempty" cbor:"-2,keyasint,omitempty"`
	BaseTime    float64  `json:"bt,omitempty" xml:"bt,attr,omitempty" cbor:"-3,keyasint,omitempty"`
	BaseUnit    string   `json:"bu,omitempty" xml:"bu,attr,omitempty" cbor:"-4,keyasint,omitempty"`
	BaseVersion uint     `json:"bver,omitempty" xml:"bver,attr,omitempty" cbor:"-1,keyasint,omitempty"`
	BaseValue   float64  `json:"bv,omitempty" xml:"bv,attr,omitempty" cbor:"-5,keyasint,omitempty"`
	BaseSum     float64  `json:"bs,omitempty" xml:"bs,attr,omitempty" cbor:"-6,keyasint,omitempty"`
	Name        string   `json:"n,omitempty" xml:"n,attr,omitempty" cbor:"0,keyasint,omitempty"`
	Unit        string   `json:"u,omitempty" xml:"u,attr,omitempty" cbor:"1,keyasint,omitempty"`
	Time        float64  `json:"t,omitempty" xml:"t,attr,omitempty" cbor:"6,keyasint,omitempty"`
	UpdateTime  float64  `json:"ut,omitempty" xml:"ut,attr,omitempty" cbor:"7,keyasint,omitempty"`
	Value       *float64 `json:"v,omitempty" xml:"v,attr,omitempty" cbor:"2,keyasint,omitempty"`
	StringValue *string  `json:"vs,omitempty" xml:"vs,attr,omitempty" cbor:"3,keyasint,omitempty"`
	DataValue   *string  `json:"vd,omitempty" xml:"vd,attr,omitempty" cbor:"8,keyasint,omitempty"`
	BoolValue   *bool    `json:"vb,omitempty" xml:"vb,attr,omitempty" cbor:"4,keyasint,omitempty"`
	Sum         *float64 `json:"s,omitempty" xml:"s,attr,omitempty" cbor:"5,keyasint,omitempty"`
}
