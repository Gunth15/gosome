package main

type Option struct {
	OptName string `json:"opt-name"`
	Desc    string `json:"desc"`
	File    string `json:"file"`
	Deps    struct {
		Package []string `json:"package"`
		Tools   []string `json:"tools"`
	} `json:"deps"`
}

func (o Option) FilterValue() string {
	return o.OptName
}

func (o Option) Title() string {
	return o.OptName
}

func (o Option) Description() string {
	return o.Desc
}

type Config struct {
	Options []Option `json:"options"`
}
