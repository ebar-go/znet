package codec

import "github.com/ebar-go/ego/utils/binary"

var defaultEndian = binary.BigEndian()

// Options represents options options
type Options struct {
	endian binary.Endian

	headerSize, headerOffset int
	actionSize, actionOffset int
	seqSize, seqOffset       int
}

func (options *Options) complete() {
	options.headerOffset = options.headerSize
	options.actionOffset = options.actionSize
	options.seqOffset = options.actionOffset + options.seqSize
}

// Default returns the default options implementation,the packet is composed by :
// |-------------- header --------------|-------- body --------|
// |packetLength| action |      seq     |-------- body --------|
// |     4      |    2   |       2      |          n           |
func DefaultOptions() *Options {
	opts := &Options{
		endian:     defaultEndian,
		headerSize: 4,
		actionSize: 2,
		seqSize:    2,
	}
	opts.complete()
	return opts
}
