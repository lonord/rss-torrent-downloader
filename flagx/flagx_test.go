package flagx

import (
	"testing"
)

func TestConvertFlag(t *testing.T) {
	got := convertFlag("TestFlag")
	want := "test_flag"
	if got != want {
		t.Errorf("convertFlag(TestFlag) = %q, want %q", got, want)
	}
}

func TestConvertFlag2(t *testing.T) {
	got := convertFlag("configFileFromX")
	want := "config_file_from_x"
	if got != want {
		t.Errorf("convertFlag(configFileFromX) = %q, want %q", got, want)
	}
}

func TestConvertFlag3(t *testing.T) {
	got := convertFlag("config-file")
	want := "config_file"
	if got != want {
		t.Errorf("convertFlag(config-file) = %q, want %q", got, want)
	}
}

func TestConvertFlag4(t *testing.T) {
	got := convertFlag("address")
	want := "address"
	if got != want {
		t.Errorf("convertFlag(address) = %q, want %q", got, want)
	}
}
