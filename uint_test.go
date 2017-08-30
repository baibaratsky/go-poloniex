package poloniex

import "testing"

func Test_convertibleUint_UnmarshalJSON(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    uint64
	}{
		{"from string", args{[]byte(`"12345"`)}, false, 12345},
		{"from number", args{[]byte(`12345`)}, false, 12345},
		{"with error", args{[]byte(`///`)}, true, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cuint := new(convertibleUint)
			err := cuint.UnmarshalJSON(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertibleUint.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}

			if uint64(*cuint) != tt.want {
				t.Errorf("convertibleUint.UnmarshalJSON() value = %d, want %d", *cuint, tt.want)
			}
		})
	}
}
