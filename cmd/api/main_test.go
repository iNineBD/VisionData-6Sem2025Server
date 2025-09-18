package main

import (
	_ "orderstreamrest/docs"
	"testing"
)

func TestMain(t *testing.T) {
	t.Skip("Adicionar testes reais")
}

func Test_main(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
		{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			main()
		})
	}
}
