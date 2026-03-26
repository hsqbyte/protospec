package checksum

// NewDefaultChecksumRegistry creates a ChecksumRegistry pre-loaded with all
// built-in checksum algorithms: internet-checksum, crc16, and crc32.
func NewDefaultChecksumRegistry() *ChecksumRegistry {
	r := NewChecksumRegistry()
	r.Register("internet-checksum", InternetChecksum)
	r.Register("crc16", CRC16)
	r.Register("crc32", CRC32)
	return r
}
