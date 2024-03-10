package validator

import (
	"testing"
)

var (
	emailRegex string = "^(([^<>()\\[\\]\\.,;:\\s@\"]+(\\.[^<>()\\[\\]\\.,;:\\s@\"]+)*)|(\".+\"))@((\\[[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}\\])|(([a-zA-Z\\-0-9]+\\.)+[a-zA-Z]{2,}))$"
)

func TestValideteByRegex(t *testing.T) {
	type args struct {
		str    string
		patern string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "standart email",
			args: args{
				str: "email@example.com",
				patern: emailRegex,
			},
			want: true,
		},
		{
			name: "username part with dot",
			args: args{
				str: "firstname.lastname@example.com",
				patern: emailRegex,
			},
			want: true,
		},
		{
			name: "username part with space domain part with dot",
			args: args{
				str: "email @subdomain.example.com",
				patern: emailRegex,
			},
			want: false,
		},
		{
			name: "username part with '+'",
			args: args{
				str: "firstname+lastname@example.com",
				patern: emailRegex,
			},
			want: true,
		},
		{
			name: "domain part with many dots",
			args: args{
				str: "email@123.123.123.123",
				patern: emailRegex,
			},
			want: false,
		},
		{
			name: "domain part with square brackets",
			args: args{
				str: "email@[123.123.123.123]",
				patern: emailRegex,
			},
			want: true,
		},
		{
			name: "username part witg quotation marks",
			args: args{
				str: "\"email\"@example.com",
				patern: emailRegex,
			},
			want: true,
		},
		{
			name: "domain part with dash",
			args: args{
				str: "email@example-one.com",
				patern: emailRegex,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValideteByRegex(tt.args.str, tt.args.patern); got != tt.want {
				t.Errorf(`ValideteByRegex(
					%s,
					%s,
					) = %v, want %v`,tt.args.str, tt.args.patern, got, tt.want)
			}
		})
	}
}
