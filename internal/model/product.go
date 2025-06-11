package model

type Product struct {
	Id         int
	Foto       string
	Titulo     string
	Descricao  string
	Preco      float64
	Quantidade int
}

type ProductDTO struct {
	Id         int     `json:"id"`
	Foto       string  `json:"foto"`
	Titulo     string  `json:"titulo"`
	Descricao  string  `json:"descricao"`
	Preco      float64 `json:"preco"`
	Quantidade int     `json:"quantidade"`
}
