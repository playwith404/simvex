package models

type Part struct {
	ID                string  `json:"id"`
	ObjectID          string  `json:"objectId"`
	Name              string  `json:"name"`
	Material          string  `json:"material"`
	Role              string  `json:"role"`
	ModelPath         string  `json:"modelPath"`
	LocalPosX         float64 `json:"localPosX"`
	LocalPosY         float64 `json:"localPosY"`
	LocalPosZ         float64 `json:"localPosZ"`
	DecomposeDirX     float64 `json:"decomposeDirX"`
	DecomposeDirY     float64 `json:"decomposeDirY"`
	DecomposeDirZ     float64 `json:"decomposeDirZ"`
	DecomposeDistance float64 `json:"decomposeDistance"`
	DecomposeOrder   int     `json:"decomposeOrder"`
}
