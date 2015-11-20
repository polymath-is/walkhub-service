package walkhub

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/lib/pq"
	"github.com/nbio/hitch"
	"github.com/tamasd/ab"
)

// AUTOGENERATED DO NOT EDIT

func NewWalkthrough(Name string) *Walkthrough {
	e := &Walkthrough{
		Name: Name,
	}

	// HOOK: newWalkthrough()

	return e
}

func EmptyWalkthrough() *Walkthrough {
	return &Walkthrough{}
}

var _ ab.Validator = &Walkthrough{}

func (e *Walkthrough) Validate() error {
	var err error

	err = validateWalkthrough(e)

	return err
}

func (e *Walkthrough) GetID() string {
	return e.UUID
}

var WalkthroughNotFoundError = errors.New("walkthrough not found")

const walkthroughFields = "w.revision, w.uuid, w.uid, w.name, w.description, w.severity, w.steps, w.updated, w.published"

func selectWalkthroughFromQuery(db ab.DB, query string, args ...interface{}) ([]*Walkthrough, error) {
	// HOOK: beforeWalkthroughSelect()

	entities := []*Walkthrough{}

	rows, err := db.Query(query, args...)

	if err != nil {
		return entities, err
	}

	for rows.Next() {
		e := EmptyWalkthrough()
		jsonSteps := ""

		if err = rows.Scan(&e.Revision, &e.UUID, &e.UID, &e.Name, &e.Description, &e.Severity, &jsonSteps, &e.Updated, &e.Published); err != nil {
			return []*Walkthrough{}, err
		}

		json.Unmarshal([]byte(jsonSteps), &e.Steps)

		entities = append(entities, e)
	}

	// HOOK: afterWalkthroughSelect()

	return entities, err
}

func selectSingleWalkthroughFromQuery(db ab.DB, query string, args ...interface{}) (*Walkthrough, error) {
	entities, err := selectWalkthroughFromQuery(db, query, args...)
	if err != nil {
		return nil, err
	}

	if len(entities) > 0 {
		return entities[0], nil
	}

	return nil, nil
}

func (e *Walkthrough) Insert(db ab.DB) error {
	beforeWalkthroughInsert(e)

	jsonSteps := ""

	bjsonSteps, _ := json.Marshal(e.Steps)
	jsonSteps = string(bjsonSteps)
	err := db.QueryRow("INSERT INTO \"walkthrough\"(uuid, uid, name, description, severity, steps, updated, published) VALUES($1, $2, $3, $4, $5, $6, $7, $8) RETURNING revision", e.UUID, e.UID, e.Name, e.Description, e.Severity, jsonSteps, e.Updated, e.Published).Scan(&e.Revision)

	// HOOK: afterWalkthroughInsert()

	return err
}

func LoadWalkthrough(db ab.DB, Revision string) (*Walkthrough, error) {
	// HOOK: beforeWalkthroughLoad()

	e, err := selectSingleWalkthroughFromQuery(db, "SELECT "+walkthroughFields+" FROM \"walkthrough\" w WHERE w.revision = $1", Revision)

	// HOOK: afterWalkthroughLoad()

	return e, err
}

func LoadAllWalkthrough(db ab.DB, start, limit int) ([]*Walkthrough, error) {
	// HOOK: beforeWalkthroughLoadAll()

	entities, err := selectWalkthroughFromQuery(db, "SELECT "+walkthroughFields+" FROM \"walkthrough\" w ORDER BY Revision DESC LIMIT $1 OFFSET $2", limit, start)

	// HOOK: afterWalkthroughLoadAll()

	return entities, err
}

func (s *WalkthroughService) Register(h *hitch.Hitch) error {
	var err error

	listMiddlewares := []func(http.Handler) http.Handler{}

	postMiddlewares := []func(http.Handler) http.Handler{}

	getMiddlewares := []func(http.Handler) http.Handler{}

	putMiddlewares := []func(http.Handler) http.Handler{}

	deleteMiddlewares := []func(http.Handler) http.Handler{}

	listMiddlewares, postMiddlewares, getMiddlewares, putMiddlewares, deleteMiddlewares = beforeWalkthroughServiceRegister()

	if err != nil {
		return err
	}

	h.Get("/api/walkthrough", s.walkthroughListHandler(), listMiddlewares...)

	h.Post("/api/walkthrough", s.walkthroughPostHandler(), postMiddlewares...)

	h.Get("/api/walkthrough/:id", s.walkthroughGetHandler(), getMiddlewares...)

	h.Put("/api/walkthrough/:id", s.walkthroughPutHandler(), putMiddlewares...)

	h.Delete("/api/walkthrough/:id", s.walkthroughDeleteHandler(), deleteMiddlewares...)

	afterWalkthroughServiceRegister(s, h)

	return err
}

func walkthroughDBErrorConverter(err *pq.Error) ab.VerboseError {
	ve := ab.NewVerboseError(err.Message, err.Detail)

	// HOOK: convertWalkthroughDBError()

	return ve
}

func (s *WalkthroughService) walkthroughListHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		db := ab.GetDB(r)
		loadFunc := LoadAllWalkthrough
		abort := false
		start := 0
		limit := 25
		if page := r.URL.Query().Get("page"); page != "" {
			pagenum, err := strconv.Atoi(page)
			ab.MaybeFail(r, http.StatusBadRequest, err)
			start = (pagenum - 1) * limit
		}

		loadFunc = beforeWalkthroughListHandler()

		if abort {
			return
		}

		entities, err := loadFunc(db, start, limit)
		ab.MaybeFail(r, http.StatusInternalServerError, ab.ConvertDBError(err, walkthroughDBErrorConverter))

		// HOOK: afterWalkthroughListHandler()

		if abort {
			return
		}

		ab.Render(r).JSON(entities)
	})
}

func (s *WalkthroughService) walkthroughPostHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		entity := &Walkthrough{}
		ab.MustDecode(r, entity)

		abort := false

		walkthroughPostValidation(entity, r)

		if abort {
			return
		}

		if err := entity.Validate(); err != nil {
			ab.Fail(r, http.StatusBadRequest, err)
		}

		db := ab.GetDB(r)

		err := entity.Insert(db)
		ab.MaybeFail(r, http.StatusInternalServerError, ab.ConvertDBError(err, walkthroughDBErrorConverter))

		afterWalkthroughPostInsertHandler(db, s, entity)

		if abort {
			return
		}

		ab.Render(r).SetCode(http.StatusCreated).JSON(entity)
	})
}

func (s *WalkthroughService) walkthroughGetHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := hitch.Params(r).ByName("id")
		db := ab.GetDB(r)
		abort := false
		loadFunc := LoadWalkthrough

		loadFunc = beforeWalkthroughGetHandler()

		if abort {
			return
		}

		entity, err := loadFunc(db, id)
		ab.MaybeFail(r, http.StatusInternalServerError, ab.ConvertDBError(err, walkthroughDBErrorConverter))
		if entity == nil {
			ab.Fail(r, http.StatusNotFound, nil)
		}

		// HOOK: afterWalkthroughGetHandler()

		if abort {
			return
		}

		ab.Render(r).JSON(entity)
	})
}

func (s *WalkthroughService) walkthroughPutHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := hitch.Params(r).ByName("id")

		entity := &Walkthrough{}
		ab.MustDecode(r, entity)

		if err := entity.Validate(); entity.UUID != id || err != nil {
			ab.Fail(r, http.StatusBadRequest, err)
		}

		db := ab.GetDB(r)
		abort := false

		beforeWalkthroughPutUpdateHandler(r, entity, db)

		if abort {
			return
		}

		err := entity.Update(db)
		ab.MaybeFail(r, http.StatusInternalServerError, ab.ConvertDBError(err, walkthroughDBErrorConverter))

		afterWalkthroughPutUpdateHandler(s, entity)

		if abort {
			return
		}

		ab.Render(r).JSON(entity)
	})
}

func (s *WalkthroughService) walkthroughDeleteHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := hitch.Params(r).ByName("id")
		db := ab.GetDB(r)
		abort := false
		loadFunc := LoadWalkthrough

		loadFunc = beforeWalkthroughDeleteHandler()

		if abort {
			return
		}

		entity, err := loadFunc(db, id)
		ab.MaybeFail(r, http.StatusInternalServerError, ab.ConvertDBError(err, walkthroughDBErrorConverter))
		if entity == nil {
			ab.Fail(r, http.StatusNotFound, nil)
		}

		insideWalkthroughDeleteHandler(r, entity, db)

		if abort {
			return
		}

		err = entity.Delete(db)
		ab.MaybeFail(r, http.StatusInternalServerError, ab.ConvertDBError(err, walkthroughDBErrorConverter))

		// HOOK: afterWalkthroughDeleteHandler()

		if abort {
			return
		}
	})
}

func (s *WalkthroughService) SchemaInstalled(db ab.DB) bool {
	found := ab.TableExists(db, "walkthrough")

	// HOOK: afterWalkthroughSchemaInstalled()

	return found
}

func (s *WalkthroughService) SchemaSQL() string {
	sql := "CREATE TABLE \"walkthrough\" (\n" +
		"\t\"revision\" uuid DEFAULT uuid_generate_v4() NOT NULL,\n" +
		"\t\"uuid\" uuid DEFAULT uuid_generate_v4() NOT NULL,\n" +
		"\t\"uid\" uuid NOT NULL,\n" +
		"\t\"name\" character varying NOT NULL,\n" +
		"\t\"description\" text NOT NULL,\n" +
		"\t\"severity\" walkthrough_severity DEFAULT 'tour' NOT NULL,\n" +
		"\t\"steps\" jsonb NOT NULL,\n" +
		"\t\"updated\" timestamp with time zone NOT NULL,\n" +
		"\t\"published\" bool NOT NULL,\n" +
		"\tCONSTRAINT walkthrough_pkey PRIMARY KEY (revision)\n);\n"

	sql = afterWalkthroughSchemaSQL(sql)

	return sql
}
