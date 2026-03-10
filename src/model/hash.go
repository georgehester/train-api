package model

type HashRequest struct {
	Password string `json:"password"`
} // @name HashRequest

type HashResponse struct {
	Hash string `json:"hash"`
} // @name HashResponse
