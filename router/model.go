package router

type Route struct {
	Key      string   `bson:"key"`
	Tipo     string   `bson:"tipo"`
	Destinos []string `bson:"destinos"`
}
