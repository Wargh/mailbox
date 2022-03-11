package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractTypeYearMonth(t *testing.T) {
	tests := []struct {
		in        string
		emailType string
		yearMonth string
		err       error
	}{
		{"inbox#2021-01", "inbox", "2021-01", nil},
		{"sent#2021-01", "sent", "2021-01", nil},
		{"sent#2021-02", "sent", "2021-02", nil},
		{"sent#2021-03", "sent", "2021-03", nil},
		{"sent#2021-09", "sent", "2021-09", nil},
		{"sent#2021-10", "sent", "2021-10", nil},
		{"sent#2021-11", "sent", "2021-11", nil},
		{"sent#2021-12", "sent", "2021-12", nil},
		// invalid
		{"sent#999-01", "", "", ErrInvalidEmailType},
		{"sent#2021-00", "", "", ErrInvalidEmailType},
		{"sent#2021-13", "", "", ErrInvalidEmailType},
		{"invalid#2021-01", "", "", ErrInvalidEmailType},
	}

	for _, test := range tests {
		emailType, yearMonth, err := ExtractTypeYearMonth(test.in)
		assert.Equal(t, test.emailType, emailType)
		assert.Equal(t, test.yearMonth, yearMonth)
		assert.Equal(t, test.err, err)
	}
}
