package noip_test

import (
	"testing"

	"mio9/dddns/internal/noip"
)

func TestRelativeRecordName(t *testing.T) {
	tests := []struct {
		zoneName   string
		recordName string
		want       string
		wantErr    bool
	}{
		{zoneName: "example.com", recordName: "home.example.com", want: "home"},
		{zoneName: "example.com", recordName: "home", want: "home"},
		{zoneName: "example.com", recordName: "example.com", want: "@"},
		{zoneName: "example.com", recordName: "@", want: "@"},
		{zoneName: "example.com", recordName: "other.com", wantErr: true},
	}

	for _, test := range tests {
		got, err := noip.RelativeRecordName(test.zoneName, test.recordName)
		if test.wantErr {
			if err == nil {
				t.Fatalf("RelativeRecordName(%q, %q) expected error", test.zoneName, test.recordName)
			}
			continue
		}
		if err != nil {
			t.Fatalf("RelativeRecordName(%q, %q): %v", test.zoneName, test.recordName, err)
		}
		if got != test.want {
			t.Fatalf("RelativeRecordName(%q, %q) = %q, want %q", test.zoneName, test.recordName, got, test.want)
		}
	}
}
