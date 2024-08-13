package index

const IndexFileName string = "index"
const _NumFilesBytesLen int = 4

// File metadata like ctime, mtime etc
const _IndexEntryMetadataLen int = 62

const _IndexEntryPaddingBytes int = 8
const _32BitToByte = 32 / 8

var _IndexFileHeader [4]byte = [4]byte{0x44, 0x49, 0x52, 0x43}           // DIRC
var _IndexFileSupportedVersion [4]byte = [4]byte{0x00, 0x00, 0x00, 0x02} // Version 2

var _TreeExtensionHeader [4]byte = [4]byte{0x54, 0x52, 0x45, 0x45} // TREE
const _TreeExtensionSize int = 4                                   // 4 bytes reserved for tree size
