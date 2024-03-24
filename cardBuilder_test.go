package izapple2

import "testing"

func TestCardBuilder(t *testing.T) {
	cardFactory := getCardFactory()
	for name, builder := range cardFactory {
		if name != "prodosromdrive" && name != "prodosromcard3" && name != "prodosnvramdrive" {
			t.Run(name, func(t *testing.T) {
				_, err := builder.buildFunc(builder.fullDefaultParams())
				if err != nil {
					t.Errorf("Exception building card '%s': %s", name, err)
				}
			})
		}
	}
}
