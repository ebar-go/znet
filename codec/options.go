package codec

import (
	"github.com/ebar-go/ego/utils/binary"
	"sync"
)

// Options represents codec options
type Options struct {
	endian binary.Endian

	headerSize, headerOffset             int
	packetLengthSize, packetLengthOffset int
	operateSize, operateOffset           int
	seqSize, seqOffset                   int
	optionSize, optionOffset             int
}

func (options *Options) complete() {
	options.endian = binary.BigEndian()
	options.headerOffset = options.headerSize
	options.packetLengthOffset = 0 + options.packetLengthSize
	options.operateOffset = options.packetLengthOffset + options.operateSize
	options.seqOffset = options.operateOffset + options.seqSize
	options.optionOffset = options.seqOffset + options.optionSize
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
		headerSize:       10,
		packetLengthSize: 4,
		operateSize:      2,
		seqSize:          2,
		optionSize:       2,
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
