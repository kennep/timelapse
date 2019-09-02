package mongo_repository

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/options"
	"github.com/mongodb/mongo-go-driver/mongo/readpref"

	"github.com/kennep/timelapse/domain"

	log "github.com/sirupsen/logrus"
)

// MongoRepository represent the repository connection
type MongoRepository struct {
	client           *mongo.Client
	database         *mongo.Database
	connectionString string
}

// NewRepository initializes the repository and connects to the database
func NewMongoRepository() (*MongoRepository, error) {
	var repository MongoRepository

	connectionString := os.Getenv("TIMELAPSE_DB_URL")
	if connectionString == "" {
		panic("Configuration error: connection string to database must be specified in TIMELAPSE_DB_URL environment variable")
	}
	database := os.Getenv("TIMELAPSE_DB_NAME")
	if database == "" {
		panic("Configuration error: database name must be specified in TIMELAPSE_DB_NAME environment variable")
	}

	authOptions := options.Client()

	username := os.Getenv("TIMELAPSE_DB_USERNAME")
	password := os.Getenv("TIMELAPSE_DB_PASSWORD")

	if username != "" && password != "" {
		credentials := options.Credential{
			Username: username,
			Password: password,
		}
		authOptions.SetAuth(credentials)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, connectionString, authOptions)
	if err != nil {
		return nil, err
	}

	ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err = client.Ping(ctx, readpref.Primary())

	repository.client = client
	repository.connectionString = client.ConnectionString()
	repository.database = client.Database(database)

	return &repository, nil
}

func (r *MongoRepository) CreateUserFromContext(appCtx *domain.ApplicationContext) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	query := bson.M{"identities": bson.M{"$elemMatch": bson.M{"subjectid": appCtx.User.SubjectID, "issuer": appCtx.User.Issuer}}}

	log.Infof("Searching for user with query: %v", query)

	result := r.database.Collection("users").FindOne(ctx, query)
	if result.Err() == nil {
		log.Info("Successfully called FindOne")
		var repoUser user
		err := result.Decode(&repoUser)
		if err == nil {
			for _, identity := range repoUser.Identities {
				if identity.SubjectID == appCtx.User.SubjectID &&
					identity.Issuer == appCtx.User.Issuer &&
					identity.Email != appCtx.User.Email {
					log.Info("Email mismatch, updating")
					identity.Email = appCtx.User.Email
					_, err := r.database.Collection("users").ReplaceOne(ctx, query, &repoUser)
					if err != nil {
						return nil, err
					}
					break
				}
			}

			domainUser := mapUserToDomain(&repoUser)
			log.Infof("Returning user: %v", domainUser)
			return domainUser, nil
		} else if err == mongo.ErrNoDocuments {
			log.Info("No documents found, inserting")
			repoUser := user{
				ID: newID(),
				Identities: []identity{{
					Issuer:    appCtx.User.Issuer,
					SubjectID: appCtx.User.SubjectID,
					Email:     appCtx.User.Email,
				}},
			}
			_, err := r.database.Collection("users").InsertOne(ctx, &repoUser)
			if err != nil {
				return nil, err
			}
			domainUser := mapUserToDomain(&repoUser)
			log.Infof("Returning user: %v", domainUser)
			return domainUser, nil
		} else {
			return nil, err
		}
	} else {
		return nil, result.Err()
	}

}

func (r *MongoRepository) AddProject(p domain.Project) (*domain.Project, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	repoProject, err := mapProjectFromDomain(&p)
	if err != nil {
		return nil, err
	}
	repoProject.ID = newID()
	_, err = r.database.Collection("projects").InsertOne(ctx, repoProject)
	if err != nil {
		return nil, err
	}

	return mapProjectToDomain(repoProject, p.User), nil
}

func (r *MongoRepository) UpdateProject(p domain.Project) (*domain.Project, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	projectid, err := stringToID(p.ID)
	if err != err {
		return nil, err
	}
	repoProject, err := mapProjectFromDomain(&p)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": projectid}

	result, err := r.database.Collection("projects").ReplaceOne(ctx, filter, repoProject)

	if err != nil {
		return nil, err
	}
	if result.MatchedCount != 1 {
		return nil, fmt.Errorf("Project with id %s not found", projectid)
	}

	return mapProjectToDomain(repoProject, p.User), nil
}

func (r *MongoRepository) GetProject(u *domain.User, projectName string) (*domain.Project, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	userid, err := stringToID(u.ID)
	if err != err {
		return nil, err
	}
	filter := bson.M{"userid": userid, "name": projectName}

	result := r.database.Collection("projects").FindOne(ctx, filter)
	if result.Err() != nil {
		return nil, err
	}

	var repoProject project
	err = result.Decode(&repoProject)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return mapProjectToDomain(&repoProject, u), nil
}

func (r *MongoRepository) getProjectById(u *domain.User, projectID primitive.ObjectID) (*domain.Project, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	userid, err := stringToID(u.ID)
	if err != err {
		return nil, err
	}
	filter := bson.M{"userid": userid, "_id": projectID}

	result := r.database.Collection("projects").FindOne(ctx, filter)
	if result.Err() != nil {
		return nil, err
	}

	var repoProject project
	err = result.Decode(&repoProject)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return mapProjectToDomain(&repoProject, u), nil
}

func (r *MongoRepository) GetProjects(u *domain.User) ([]*domain.Project, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	userid, err := stringToID(u.ID)
	if err != err {
		return nil, err
	}
	filter := bson.M{"userid": userid}

	var result []*domain.Project

	cursor, err := r.database.Collection("projects").Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	for cursor.Next(ctx) {
		var repoProject project
		err = cursor.Decode(&repoProject)
		if err != nil {
			return nil, err
		}
		result = append(result, mapProjectToDomain(&repoProject, u))
	}
	if err = cursor.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *MongoRepository) AddTimeEntry(p *domain.Project, e domain.TimeEntry) (*domain.TimeEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	e.Project = p

	repoEntry, err := mapTimeEntryFromDomain(&e)
	repoEntry.ID = newID()
	if err != nil {
		return nil, err
	}
	_, err = r.database.Collection("timeentries").InsertOne(ctx, repoEntry)
	if err != nil {
		return nil, err
	}

	e.ID = idToString(repoEntry.ID)
	return &e, nil
}

func (r *MongoRepository) UpdateTimeEntry(p *domain.Project, e domain.TimeEntry) (*domain.TimeEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	repoEntry, err := mapTimeEntryFromDomain(&e)
	if err != nil {
		return nil, err
	}

	filter := bson.M{"_id": repoEntry.ID}

	result, err := r.database.Collection("timeentries").ReplaceOne(ctx, filter, repoEntry)

	if err != nil {
		return nil, err
	}
	if result.MatchedCount != 1 {
		return nil, fmt.Errorf("Time entry with id %s not found", e.ID)
	}

	return mapTimeEntryToDomain(repoEntry, p), nil
}

func (r *MongoRepository) GetProjectTimeEntries(p *domain.Project) ([]*domain.TimeEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	projectid, err := stringToID(p.ID)
	if err != err {
		return nil, err
	}
	filter := bson.M{"projectid": projectid}

	var result []*domain.TimeEntry

	cursor, err := r.database.Collection("timeentries").Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	for cursor.Next(ctx) {
		var repoEntry timeEntry
		err = cursor.Decode(&repoEntry)
		if err != nil {
			return nil, err
		}
		result = append(result, mapTimeEntryToDomain(&repoEntry, p))
	}
	if err = cursor.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *MongoRepository) GetProjectTimeEntry(p *domain.Project, entryID string) (*domain.TimeEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	projectid, err := stringToID(p.ID)
	if err != err {
		return nil, err
	}
	entryid, err := stringToID(entryID)
	if err != err {
		return nil, err
	}

	filter := bson.M{
		"_id":       entryid,
		"projectid": projectid,
	}

	result := r.database.Collection("timeentries").FindOne(ctx, filter)
	if result.Err() != nil {
		return nil, err
	}
	var repoEntry timeEntry
	err = result.Decode(&repoEntry)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return mapTimeEntryToDomain(&repoEntry, p), nil
}

func (r *MongoRepository) GetUserTimeEntries(u *domain.User) ([]*domain.TimeEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	userid, err := stringToID(u.ID)
	if err != err {
		return nil, err
	}
	filter := bson.M{"userid": userid}

	var result []*domain.TimeEntry

	cursor, err := r.database.Collection("timeentries").Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	for cursor.Next(ctx) {
		var repoEntry timeEntry
		err = cursor.Decode(&repoEntry)
		if err != nil {
			return nil, err
		}

		project, err := r.getProjectById(u, repoEntry.ProjectID)
		if err != nil {
			return nil, err
		}
		result = append(result, mapTimeEntryToDomain(&repoEntry, project))
	}
	if err = cursor.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
