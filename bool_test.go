package poloniex

import "testing"

func Test_convertibleBool_UnmarshalJSON(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    bool
	}{
		{"true number value", args{[]byte(`"1"`)}, false, true},
		{"true boolean value", args{[]byte(`"true"`)}, false, true},
		{"false number value", args{[]byte(`"0"`)}, false, false},
		{"false boolean value", args{[]byte(`"false"`)}, false, false},
		{"error value", args{[]byte(`"//"`)}, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cbool := new(convertibleBool)
			err := cbool.UnmarshalJSON(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertibleBool.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}

			if bool(*cbool) != tt.want {
				t.Errorf("convertibleBool.UnmarshalJSON() value = %v, want %v", bool(*cbool), tt.want)
			}
		})
	}
}
