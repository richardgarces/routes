package router

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Repository interface {
	GetRoute(key, tipo string) (*Route, error)
	SaveRoute(key, tipo, destino string) error
	GetAllRoutes() ([]Route, error)
}

type repo struct {
	col *mongo.Collection
}

func NewRepository(db *mongo.Database) Repository {
	return &repo{col: db.Collection("routes")}
}

func (r *repo) GetRoute(key, tipo string) (*Route, error) {
	log.Printf("[Repository] Consulta a MongoDB: key='%s', tipo='%s'", key, tipo)
	var route Route
	err := r.col.FindOne(context.TODO(), bson.M{"key": key, "tipo": tipo}).Decode(&route)
	if err != nil {
		log.Printf("[Repository] Error en FindOne: %v", err)
		return nil, err
	}
	log.Printf("[Repository] Documento encontrado: %+v", route)
	return &route, nil
}

func (r *repo) SaveRoute(key, tipo, destino string) error {
	log.Printf("Guardando destino %s en la key %s, tipo %s en la base de datos", destino, key, tipo)
	_, err := r.col.UpdateOne(context.TODO(),
		bson.M{"key": key, "tipo": tipo},
		bson.M{"$addToSet": bson.M{"destinos": destino}},
	)
	return err
}

func (r *repo) GetAllRoutes() ([]Route, error) {
	log.Println("Obteniendo todas las rutas de la base de datos...")
	cursor, err := r.col.Find(context.TODO(), bson.M{})
	if err != nil {
		log.Printf("Error al obtener rutas: %v", err)
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var routes []Route
	for cursor.Next(context.TODO()) {
		var route Route
		if err := cursor.Decode(&route); err != nil {
			log.Printf("Error decodificando ruta: %v", err)
			continue
		}
		routes = append(routes, route)
	}
	return routes, nil
}
