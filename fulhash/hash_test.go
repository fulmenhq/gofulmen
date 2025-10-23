package fulhash

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fulmenhq/gofulmen/schema"
	"gopkg.in/yaml.v3"
)

type FixtureFile struct {
	Version           string             `yaml:"version"`
	Description       string             `yaml:"description"`
	Fixtures          []BlockFixture     `yaml:"fixtures"`
	StreamingFixtures []StreamingFixture `yaml:"streaming_fixtures"`
	ErrorFixtures     []ErrorFixture     `yaml:"error_fixtures"`
	FormatFixtures    []FormatFixture    `yaml:"format_fixtures"`
}

type BlockFixture struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Input       string `yaml:"input,omitempty"`
	Encoding    string `yaml:"encoding"`
	InputBytes  []int  `yaml:"input_bytes,omitempty"`
	XXH3_128    string `yaml:"xxh3_128"`
	SHA256      string `yaml:"sha256"`
	Notes       string `yaml:"notes,omitempty"`
}

type StreamingFixture struct {
	Name             string  `yaml:"name"`
	Description      string  `yaml:"description"`
	Chunks           []Chunk `yaml:"chunks"`
	ExpectedXXH3_128 string  `yaml:"expected_xxh3_128"`
	ExpectedSHA256   string  `yaml:"expected_sha256"`
	Notes            string  `yaml:"notes,omitempty"`
}

type Chunk struct {
	Value    string `yaml:"value,omitempty"`
	Encoding string `yaml:"encoding,omitempty"`
	Size     int    `yaml:"size,omitempty"`
	Pattern  string `yaml:"pattern,omitempty"`
}

type ErrorFixture struct {
	Name                 string   `yaml:"name"`
	Input                string   `yaml:"input,omitempty"`
	Algorithm            string   `yaml:"algorithm,omitempty"`
	Checksum             string   `yaml:"checksum,omitempty"`
	ExpectedError        string   `yaml:"expected_error"`
	ErrorMessageContains []string `yaml:"error_message_contains,omitempty"`
}

type FormatFixture struct {
	Name              string `yaml:"name"`
	Algorithm         string `yaml:"algorithm,omitempty"`
	Hex               string `yaml:"hex,omitempty"`
	Formatted         string `yaml:"formatted,omitempty"`
	ExpectedFormatted string `yaml:"expected_formatted,omitempty"`
	ExpectedAlgorithm string `yaml:"expected_algorithm,omitempty"`
	ExpectedHex       string `yaml:"expected_hex,omitempty"`
}

func loadFixtures(t *testing.T) *FixtureFile {
	t.Helper()
	fixturePath := "../config/crucible-go/library/fulhash/fixtures.yaml"
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("Failed to read fixtures: %v", err)
	}

	var fixtures FixtureFile
	if err := yaml.Unmarshal(data, &fixtures); err != nil {
		t.Fatalf("Failed to parse fixtures: %v", err)
	}
	return &fixtures
}

func TestHashString(t *testing.T) {
	fixtures := loadFixtures(t)

	// Find hello-world fixture
	var fixture *BlockFixture
	for i := range fixtures.Fixtures {
		if fixtures.Fixtures[i].Name == "hello-world" {
			fixture = &fixtures.Fixtures[i]
			break
		}
	}
	if fixture == nil {
		t.Fatal("hello-world fixture not found")
	}

	digest, err := HashString(fixture.Input, WithAlgorithm(XXH3_128))
	if err != nil {
		t.Fatalf("HashString failed: %v", err)
	}
	if digest.String() != fixture.XXH3_128 {
		t.Errorf("HashString mismatch: got %s, want %s", digest.String(), fixture.XXH3_128)
	}
}

func TestParseDigest(t *testing.T) {
	fixtures := loadFixtures(t)

	tests := []struct {
		name    string
		input   string
		wantAlg Algorithm
		wantHex string
		wantErr bool
	}{
		{"valid-xxh3", "xxh3-128:abc123", XXH3_128, "abc123", false},
		{"valid-sha256", "sha256:def456", SHA256, "def456", false},
		{"invalid-format", "invalid", "", "", true},
		{"unknown-algorithm", "unknown:abc", "", "", true},
		{"invalid-hex", "xxh3-128:invalidhex", "", "", true},
	}

	// Add tests from format fixtures
	for _, ff := range fixtures.FormatFixtures {
		if ff.Formatted != "" && ff.ExpectedAlgorithm != "" && ff.ExpectedHex != "" {
			tests = append(tests, struct {
				name    string
				input   string
				wantAlg Algorithm
				wantHex string
				wantErr bool
			}{
				name:    "fixture-" + ff.Name,
				input:   ff.Formatted,
				wantAlg: Algorithm(ff.ExpectedAlgorithm),
				wantHex: ff.ExpectedHex,
				wantErr: false,
			})
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := ParseDigest(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDigest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if d.Algorithm() != tt.wantAlg {
					t.Errorf("Algorithm = %v, want %v", d.Algorithm(), tt.wantAlg)
				}
				if d.Hex() != tt.wantHex {
					t.Errorf("Hex = %v, want %v", d.Hex(), tt.wantHex)
				}
			}
		})
	}
}

func TestHasher(t *testing.T) {
	data := []byte("Hello, World!")

	// Test XXH3-128
	hasher, err := NewHasher(WithAlgorithm(XXH3_128))
	if err != nil {
		t.Fatalf("NewHasher failed: %v", err)
	}
	_, err = hasher.Write(data)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	digest := hasher.Sum()
	expected := "xxh3-128:531df2844447dd5077db03842cd75395"
	if digest.String() != expected {
		t.Errorf("XXH3-128 hasher mismatch: got %s, want %s", digest.String(), expected)
	}

	// Test Reset
	hasher.Reset()
	_, err = hasher.Write([]byte("test"))
	if err != nil {
		t.Fatalf("Write after reset failed: %v", err)
	}
	digest2 := hasher.Sum()
	if digest2.String() == expected {
		t.Errorf("Reset did not work: got same digest %s", digest2.String())
	}

	// Test SHA256 hasher
	hasher256, err := NewHasher(WithAlgorithm(SHA256))
	if err != nil {
		t.Fatalf("NewHasher SHA256 failed: %v", err)
	}
	_, err = hasher256.Write(data)
	if err != nil {
		t.Fatalf("SHA256 Write failed: %v", err)
	}
	digest256 := hasher256.Sum()
	if digest256.Algorithm() != SHA256 {
		t.Errorf("SHA256 hasher algorithm: got %s, want %s", digest256.Algorithm(), SHA256)
	}
	if len(digest256.Bytes()) != 32 {
		t.Errorf("SHA256 bytes length: got %d, want 32", len(digest256.Bytes()))
	}
}

func TestUnsupportedAlgorithm(t *testing.T) {
	_, err := Hash([]byte("test"), WithAlgorithm("md5"))
	if err == nil {
		t.Error("Expected error for unsupported algorithm")
	}

	_, err = NewHasher(WithAlgorithm("md5"))
	if err == nil {
		t.Error("Expected error for unsupported algorithm in NewHasher")
	}
}

func TestErrorFixtures(t *testing.T) {
	fixtures := loadFixtures(t)

	for _, ef := range fixtures.ErrorFixtures {
		t.Run("error/"+ef.Name, func(t *testing.T) {
			var err error
			switch {
			case ef.Algorithm != "":
				// Test Hash with unsupported algorithm
				_, err = Hash([]byte(ef.Input), WithAlgorithm(Algorithm(ef.Algorithm)))
			case ef.Checksum != "":
				// Test ParseDigest with invalid checksum
				_, err = ParseDigest(ef.Checksum)
			default:
				t.Fatalf("error fixture %s has no algorithm or checksum", ef.Name)
			}

			if err == nil {
				t.Errorf("expected error %s, got nil", ef.ExpectedError)
				return
			}

			// Check if error message contains expected strings
			errMsg := err.Error()
			for _, contains := range ef.ErrorMessageContains {
				if !strings.Contains(errMsg, contains) {
					t.Errorf("error message %q does not contain %q", errMsg, contains)
				}
			}
		})
	}
}

func TestHashReader(t *testing.T) {
	reader := strings.NewReader("Hello, World!")
	digest, err := HashReader(reader, WithAlgorithm(XXH3_128))
	if err != nil {
		t.Fatalf("HashReader failed: %v", err)
	}
	expected := "xxh3-128:531df2844447dd5077db03842cd75395"
	if digest.String() != expected {
		t.Errorf("HashReader mismatch: got %s, want %s", digest.String(), expected)
	}
}

func TestHash_Empty(t *testing.T) {
	digest, err := Hash([]byte{}, WithAlgorithm(XXH3_128))
	if err != nil {
		t.Fatalf("Hash empty failed: %v", err)
	}
	expected := "xxh3-128:99aa06d3014798d86001c324468d497f"
	if digest.String() != expected {
		t.Errorf("Empty hash mismatch: got %s, want %s", digest.String(), expected)
	}
}

func TestHasher_MultipleWrites(t *testing.T) {
	hasher, err := NewHasher(WithAlgorithm(XXH3_128))
	if err != nil {
		t.Fatalf("NewHasher failed: %v", err)
	}
	_, err = hasher.Write([]byte("Hello, "))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	_, err = hasher.Write([]byte("World!"))
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	digest := hasher.Sum()
	expected := "xxh3-128:531df2844447dd5077db03842cd75395"
	if digest.String() != expected {
		t.Errorf("Multiple writes mismatch: got %s, want %s", digest.String(), expected)
	}
}

func TestOptions(t *testing.T) {
	// Test default algorithm
	digest, err := Hash([]byte("test"))
	if err != nil {
		t.Fatalf("Default options failed: %v", err)
	}
	if digest.Algorithm() != XXH3_128 {
		t.Errorf("Default algorithm: got %s, want %s", digest.Algorithm(), XXH3_128)
	}

	// Test with SHA256
	digest, err = Hash([]byte("test"), WithAlgorithm(SHA256))
	if err != nil {
		t.Fatalf("SHA256 failed: %v", err)
	}
	if digest.Algorithm() != SHA256 {
		t.Errorf("SHA256 algorithm: got %s, want %s", digest.Algorithm(), SHA256)
	}
}

func TestDigest_Methods(t *testing.T) {
	digest, err := Hash([]byte("test"), WithAlgorithm(XXH3_128))
	if err != nil {
		t.Fatalf("Hash failed: %v", err)
	}

	if digest.Algorithm() != XXH3_128 {
		t.Errorf("Algorithm: got %s, want %s", digest.Algorithm(), XXH3_128)
	}

	hex := digest.Hex()
	if len(hex) != 32 { // 128 bits = 32 hex chars
		t.Errorf("Hex length: got %d, want 32", len(hex))
	}

	bytes := digest.Bytes()
	if len(bytes) != 16 {
		t.Errorf("Bytes length: got %d, want 16", len(bytes))
	}

	formatted := digest.String()
	expected := "xxh3-128:" + hex
	if formatted != expected {
		t.Errorf("String: got %s, want %s", formatted, expected)
	}
}

func TestFormatDigest(t *testing.T) {
	digest, _ := Hash([]byte("test"), WithAlgorithm(XXH3_128))
	formatted := FormatDigest(digest)
	if formatted != digest.String() {
		t.Errorf("FormatDigest: got %s, want %s", formatted, digest.String())
	}
}

func TestStreamingVsBlock(t *testing.T) {
	data := []byte("This is a test string for streaming vs block hashing.")

	// Block hash
	blockDigest, err := Hash(data, WithAlgorithm(XXH3_128))
	if err != nil {
		t.Fatalf("Block hash failed: %v", err)
	}

	// Streaming hash
	hasher, err := NewHasher(WithAlgorithm(XXH3_128))
	if err != nil {
		t.Fatalf("NewHasher failed: %v", err)
	}
	n, err := hasher.Write(data)
	if err != nil {
		t.Fatalf("Streaming write failed: %v", err)
	}
	if n != len(data) {
		t.Fatalf("Write returned wrong length: got %d, want %d", n, len(data))
	}
	streamDigest := hasher.Sum()

	if blockDigest.String() != streamDigest.String() {
		t.Errorf("Block and streaming mismatch: block %s, stream %s", blockDigest.String(), streamDigest.String())
	}
}

func TestFulHashFixturesValidation(t *testing.T) {
	fixturePath := filepath.Join("..", "config", "crucible-go", "library", "fulhash", "fixtures.yaml")

	// Validate fixtures against JSON schema
	catalog := schema.DefaultCatalog()
	validator, err := catalog.ValidatorByID("library/fulhash/v1.0.0/fixtures")
	if err != nil {
		t.Fatalf("Failed to get schema validator: %v", err)
	}

	diagnostics, err := validator.ValidateFile(fixturePath)
	if err != nil {
		t.Fatalf("Schema validation failed with error: %v", err)
	}
	if len(diagnostics) > 0 {
		t.Errorf("Fixtures violate schema:")
		for _, diag := range diagnostics {
			t.Errorf("  %s (%s): %s", diag.Pointer, diag.Keyword, diag.Message)
		}
		t.Fatalf("Schema validation failed with %d diagnostics", len(diagnostics))
	}

	// Load fixtures for additional runtime checks
	fixtures := loadFixtures(t)

	// Test all block fixtures with both algorithms
	for _, f := range fixtures.Fixtures {
		t.Run("block/"+f.Name+"/xxh3-128", func(t *testing.T) {
			var input []byte
			if f.InputBytes != nil {
				// Use input_bytes if present
				input = make([]byte, len(f.InputBytes))
				for i, b := range f.InputBytes {
					input[i] = byte(b)
				}
			} else {
				// Use input string (may be empty)
				input = []byte(f.Input)
			}

			digest, err := Hash(input, WithAlgorithm(XXH3_128))
			if err != nil {
				t.Fatalf("Hash failed for fixture %s: %v", f.Name, err)
			}

			if digest.String() != f.XXH3_128 {
				t.Errorf("fixture %s xxh3-128 mismatch: got %s, want %s", f.Name, digest.String(), f.XXH3_128)
			}
		})

		t.Run("block/"+f.Name+"/sha256", func(t *testing.T) {
			var input []byte
			if f.InputBytes != nil {
				// Use input_bytes if present
				input = make([]byte, len(f.InputBytes))
				for i, b := range f.InputBytes {
					input[i] = byte(b)
				}
			} else {
				// Use input string (may be empty)
				input = []byte(f.Input)
			}

			digest, err := Hash(input, WithAlgorithm(SHA256))
			if err != nil {
				t.Fatalf("Hash failed for fixture %s: %v", f.Name, err)
			}

			if digest.String() != f.SHA256 {
				t.Errorf("fixture %s sha256 mismatch: got %s, want %s", f.Name, digest.String(), f.SHA256)
			}
		})
	}

	// Test all streaming fixtures with both algorithms
	for _, sf := range fixtures.StreamingFixtures {
		t.Run("streaming/"+sf.Name+"/xxh3-128", func(t *testing.T) {
			hasher, err := NewHasher(WithAlgorithm(XXH3_128))
			if err != nil {
				t.Fatalf("NewHasher failed for fixture %s: %v", sf.Name, err)
			}

			for _, chunk := range sf.Chunks {
				var data []byte
				if chunk.Value != "" {
					data = []byte(chunk.Value)
				} else if chunk.Size > 0 {
					if chunk.Pattern != "" {
						// Generate pattern data - pattern like "repeating-A" means repeat 'A'
						data = make([]byte, chunk.Size)
						var patternByte byte
						switch chunk.Pattern {
						case "repeating-A":
							patternByte = 'A'
						case "repeating-B":
							patternByte = 'B'
						case "repeating-C":
							patternByte = 'C'
						default:
							t.Fatalf("unknown pattern %s in fixture %s", chunk.Pattern, sf.Name)
						}
						for i := 0; i < chunk.Size; i++ {
							data[i] = patternByte
						}
					} else {
						data = make([]byte, chunk.Size)
					}
				} else {
					t.Fatalf("chunk in fixture %s has no value, size, or pattern", sf.Name)
				}

				_, err := hasher.Write(data)
				if err != nil {
					t.Fatalf("Write failed for fixture %s: %v", sf.Name, err)
				}
			}

			digest := hasher.Sum()
			if digest.String() != sf.ExpectedXXH3_128 {
				t.Errorf("fixture %s xxh3-128 mismatch: got %s, want %s", sf.Name, digest.String(), sf.ExpectedXXH3_128)
			}
		})

		t.Run("streaming/"+sf.Name+"/sha256", func(t *testing.T) {
			hasher, err := NewHasher(WithAlgorithm(SHA256))
			if err != nil {
				t.Fatalf("NewHasher failed for fixture %s: %v", sf.Name, err)
			}

			for _, chunk := range sf.Chunks {
				var data []byte
				if chunk.Value != "" {
					data = []byte(chunk.Value)
				} else if chunk.Size > 0 {
					if chunk.Pattern != "" {
						// Generate pattern data - pattern like "repeating-A" means repeat 'A'
						data = make([]byte, chunk.Size)
						var patternByte byte
						switch chunk.Pattern {
						case "repeating-A":
							patternByte = 'A'
						case "repeating-B":
							patternByte = 'B'
						case "repeating-C":
							patternByte = 'C'
						default:
							t.Fatalf("unknown pattern %s in fixture %s", chunk.Pattern, sf.Name)
						}
						for i := 0; i < chunk.Size; i++ {
							data[i] = patternByte
						}
					} else {
						data = make([]byte, chunk.Size)
					}
				} else {
					t.Fatalf("chunk in fixture %s has no value, size, or pattern", sf.Name)
				}

				_, err := hasher.Write(data)
				if err != nil {
					t.Fatalf("Write failed for fixture %s: %v", sf.Name, err)
				}
			}

			digest := hasher.Sum()
			if digest.String() != sf.ExpectedSHA256 {
				t.Errorf("fixture %s sha256 mismatch: got %s, want %s", sf.Name, digest.String(), sf.ExpectedSHA256)
			}
		})
	}
}
