package codec

import (
	"github.com/ebar-go/ego/utils/binary"
	"sync"
)

// Options represents codec options
type Options struct {
	// ContentType is data content type
	ContentType int

	endian binary.Endian

	headerSize, headerOffset             int
	packetLengthSize, packetLengthOffset int
	operateSize, operateOffset           int
	contentTypeSize, contentTypeOffset   int
	seqSize, seqOffset                   int
}

func (options *Options) complete() {
	options.endian = binary.BigEndian()
	options.headerOffset = options.headerSize
	options.packetLengthOffset = 0 + options.packetLengthSize
	options.operateOffset = options.packetLengthOffset + options.operateSize
	options.contentTypeOffset = options.operateOffset + options.contentTypeSize
	options.seqOffset = options.contentTypeOffset + options.seqSize
}

func (options *Options) New() Codec {
	return &DefaultCodec{options: options}
}

func (options *Options) NewWithHeader(header Header) Codec {
	return &DefaultCodec{options: options, header: header}
}

type Option func(options *Options)

// Default returns the default codec implementation,the packet is composed by :
// |-------------- header ------------- |-------- body --------|
// |packetLength|operate|contentType|seq|-------- body --------|
// |     4      |   2   |      2    | 2 |          n           |
func defaultOptions() *Options {
	return &Options{
		ContentType:      ContentTypeJSON,
		headerSize:       10,
		packetLengthSize: 4,
		operateSize:      2,
		contentTypeSize:  2,
		seqSize:          2,
	}
}

var optionsInstance = struct {
	once     sync.Once
	instance *Options
}{}

func Factory() *Options {
	optionsInstance.once.Do(func() {
		optionsInstance.instance = defaultOptions()
		optionsInstance.instance.complete()
	})
	return optionsInstance.instance
}
