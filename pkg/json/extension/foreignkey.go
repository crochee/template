package extension

import (
	"unsafe"

	jsoniter "github.com/json-iterator/go"
	"gorm.io/gorm/schema"
)

type IgnoreForeignKeyCodec struct {
	jsoniter.DummyExtension
}

func (extension *IgnoreForeignKeyCodec) UpdateStructDescriptor(structDescriptor *jsoniter.StructDescriptor) {
	for _, binding := range structDescriptor.Fields {
		tag, hastag := binding.Field.Tag().Lookup("gorm")
		if hastag {
			tagSetting := schema.ParseTagSetting(tag, ";")
			if _, ok := tagSetting["FOREIGNKEY"]; ok {
				binding.Encoder = &funcEncoder{fun: func(_ unsafe.Pointer, stream *jsoniter.Stream) {
					stream.WriteNil()
				}}
			}
		}
	}
}
