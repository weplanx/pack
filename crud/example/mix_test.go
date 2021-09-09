package example

import (
	"bytes"
	"errors"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/kainonly/go-bit/crud"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

type UserMixController struct {
	*crud.Crud
}

func (x *UserMixController) FindOne(c *gin.Context) interface{} {
	var body struct {
		crud.FindOneBody
		Name string `json:"name"`
	}
	crud.Mix(c,
		crud.SetBody(&body),
		crud.Query(func(tx *gorm.DB) *gorm.DB {
			tx = tx.Where("name = ?", body.Name)
			return tx
		}),
	)
	return x.Crud.FindOne(c)
}

func TestMixFindOne(t *testing.T) {
	res := httptest.NewRecorder()
	body, _ := jsoniter.Marshal(&map[string]interface{}{
		"name": "Marcia",
	})
	req, _ := http.NewRequest("POST", "/user-mix/r/find/one", bytes.NewBuffer(body))
	r.ServeHTTP(res, req)
	assert.Equal(t,
		res.Body.String(),
		`{"data":{"id":8,"path":"Marcia@VX.com","name":"Marcia","age":37,"gender":"Female","department":"Support"},"error":0}`,
	)
}

func (x *UserMixController) FindMany(c *gin.Context) interface{} {
	crud.Mix(c,
		crud.Query(func(tx *gorm.DB) *gorm.DB {
			tx.Where("id in ?", []uint64{5, 6})
			return tx
		}),
	)
	return x.Crud.FindMany(c)
}

func TestMixFindMany(t *testing.T) {
	res := httptest.NewRecorder()
	body, _ := jsoniter.Marshal(&map[string]interface{}{
		"order": crud.Orders{
			"id": "desc",
		},
	})
	req, _ := http.NewRequest("POST", "/user-mix/r/find/many", bytes.NewBuffer(body))
	r.ServeHTTP(res, req)
	assert.Equal(t,
		res.Body.String(),
		`{"data":[{"id":6,"path":"Max@VX.com","name":"Max","age":28,"gender":"Female","department":"Designer"},{"id":5,"path":"Vivianne@VX.com","name":"Vivianne","age":36,"gender":"Male","department":"Sale"}],"error":0}`,
	)
}

func (x *UserMixController) Create(c *gin.Context) interface{} {
	crud.Mix(c,
		crud.TxNext(func(tx *gorm.DB, args ...interface{}) error {
			log.Println(args[0].(*Example))
			return errors.New("an abnormal rollback occurred")
		}),
	)
	return x.Crud.Create(c)
}

func TestMixCreate(t *testing.T) {
	res := httptest.NewRecorder()
	body, _ := jsoniter.Marshal(&Example{
		Email:      "Zhang@VX.com",
		Name:       "Zhang",
		Age:        27,
		Gender:     "Male",
		Department: "IT",
	})
	req, _ := http.NewRequest("POST", "/user-mix/w/create", bytes.NewBuffer(body))
	r.ServeHTTP(res, req)
	assert.Equal(t,
		res.Body.String(),
		`{"error":1,"msg":"an abnormal rollback occurred"}`,
	)
	var count int64
	err = db.Model(&Example{}).Where("name = ?", "Zhang").Count(&count).Error
	assert.Nil(t, err)
	assert.Equal(t, count, int64(0))
}

func (x *UserMixController) Update(c *gin.Context) interface{} {
	var body struct {
		crud.UpdateBody
		Name string `json:"name"`
	}
	crud.Mix(c,
		crud.SetBody(&body),
		crud.Query(func(tx *gorm.DB) *gorm.DB {
			tx = tx.Where("name = ?", body.Name)
			return tx
		}),
		crud.TxNext(func(tx *gorm.DB, args ...interface{}) error {
			log.Println(args[0].(*Example))
			return errors.New("an abnormal rollback occurred")
		}),
	)
	return x.Crud.Update(c)
}

func TestMixUpdate(t *testing.T) {
	res := httptest.NewRecorder()
	body, _ := jsoniter.Marshal(map[string]interface{}{
		"name":    "Stuart",
		"updates": Example{Age: 25},
	})
	req, _ := http.NewRequest("POST", "/user-mix/w/update", bytes.NewBuffer(body))
	r.ServeHTTP(res, req)
	assert.Equal(t,
		res.Body.String(),
		`{"error":1,"msg":"an abnormal rollback occurred"}`,
	)
	var data Example
	err = db.Where("name = ?", "Stuart").First(&data).Error
	assert.Nil(t, err)
	assert.Equal(t, data.Age, 27)
}

func (x *UserMixController) Delete(c *gin.Context) interface{} {
	var body struct {
		crud.DeleteBody
		Name string `json:"name"`
	}
	crud.Mix(c,
		crud.SetBody(&body),
		crud.Query(func(tx *gorm.DB) *gorm.DB {
			tx = tx.Where("name = ?", body.Name)
			return tx
		}),
		crud.TxNext(func(tx *gorm.DB, args ...interface{}) error {
			log.Println(args[0])
			return errors.New("an abnormal rollback occurred")
		}),
	)
	return x.Crud.Delete(c)
}

func TestMixDelete(t *testing.T) {
	res := httptest.NewRecorder()
	body, _ := jsoniter.Marshal(map[string]interface{}{
		"name": "Joanna",
	})
	req, _ := http.NewRequest("POST", "/user-mix/w/delete", bytes.NewBuffer(body))
	r.ServeHTTP(res, req)
	assert.Equal(t,
		res.Body.String(),
		`{"error":1,"msg":"an abnormal rollback occurred"}`,
	)
	var count int64
	err = db.Model(&Example{}).Where("name = ?", "Joanna").Count(&count).Error
	assert.Nil(t, err)
	assert.Equal(t, count, int64(1))
}
