package block_test

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/numbatx/gn-numbat/api/block"
	apiErrors "github.com/numbatx/gn-numbat/api/errors"
	"github.com/numbatx/gn-numbat/api/mock"
	"github.com/numbatx/gn-numbat/api/node"
	"github.com/numbatx/gn-numbat/node/external"
	"github.com/stretchr/testify/assert"
)

type errorResponse struct {
	Error string `json:"error"`
}

type recentBlocksResponse struct {
	errorResponse
	Blocks      []*external.BlockHeader `json:"blocks"`
	ShardHeader *external.BlockHeader   `json:"block"`
}

func init() {
	gin.SetMode(gin.TestMode)
}

func loadResponse(rsp io.Reader, destination interface{}) {
	jsonParser := json.NewDecoder(rsp)
	err := jsonParser.Decode(destination)
	if err != nil {
		logError(err)
	}
}

func logError(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func startNodeServer(handler node.FacadeHandler) *gin.Engine {
	server := startNodeServerWithFacade(handler)
	return server
}

func startNodeServerWrongFacade() *gin.Engine {
	return startNodeServerWithFacade(mock.WrongFacade{})
}

func startNodeServerWithFacade(facade interface{}) *gin.Engine {
	ws := gin.New()
	ws.Use(cors.Default())
	if facade != nil {
		ws.Use(func(c *gin.Context) {
			c.Set("numbatFacade", facade)
		})
	}

	blockRoutes := ws.Group("/block")
	block.Routes(blockRoutes)
	blocksRoutes := ws.Group("/blocks")
	block.RoutesForBlocksLists(blocksRoutes)
	return ws
}

//------- RecentBlocks

func TestRecentBlocks_FailsWithoutFacade(t *testing.T) {
	t.Parallel()
	ws := startNodeServer(nil)
	defer func() {
		r := recover()
		assert.NotNil(t, r, "Not providing numbatFacade context should panic")
	}()
	req, _ := http.NewRequest("GET", "/blocks/recent", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)
}

func TestRecentBlocks_FailsWithWrongFacadeTypeConversion(t *testing.T) {
	t.Parallel()
	ws := startNodeServerWrongFacade()
	req, _ := http.NewRequest("GET", "/blocks/recent", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	statusRsp := errorResponse{}
	loadResponse(resp.Body, &statusRsp)
	assert.Equal(t, resp.Code, http.StatusInternalServerError)
	assert.Equal(t, statusRsp.Error, apiErrors.ErrInvalidAppContext.Error())
}

func TestRecentBlocks_ReturnsCorrectly(t *testing.T) {
	t.Parallel()
	recentBlocks := []*external.BlockHeader{
		{Nonce: 0, Hash: make([]byte, 0), PrevHash: make([]byte, 0), StateRootHash: make([]byte, 0)},
		{Nonce: 0, Hash: make([]byte, 0), PrevHash: make([]byte, 0), StateRootHash: make([]byte, 0)},
	}
	facade := mock.Facade{
		RecentNotarizedBlocksHandler: func(maxShardHeadersNum int) (blocks []*external.BlockHeader, e error) {
			return recentBlocks, nil
		},
	}

	ws := startNodeServer(&facade)
	req, _ := http.NewRequest("GET", "/blocks/recent", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	rb := recentBlocksResponse{}
	loadResponse(resp.Body, &rb)
	assert.Equal(t, resp.Code, http.StatusOK)
	assert.NotNil(t, rb.Blocks)
	assert.Equal(t, recentBlocks, rb.Blocks)
}

func TestRecentBlocks_ReturnsErrorWhenRecentBlocksErrors(t *testing.T) {
	t.Parallel()
	errMessage := "recent blocks error"
	facade := mock.Facade{
		RecentNotarizedBlocksHandler: func(maxShardHeadersNum int) (blocks []*external.BlockHeader, e error) {
			return nil, errors.New(errMessage)
		},
	}

	ws := startNodeServer(&facade)
	req, _ := http.NewRequest("GET", "/blocks/recent", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	rb := recentBlocksResponse{}
	loadResponse(resp.Body, &rb)
	assert.Equal(t, resp.Code, http.StatusInternalServerError)
	assert.Nil(t, rb.Blocks)
	assert.NotNil(t, rb.Error)
	assert.Equal(t, errMessage, rb.Error)
}

//------- Block

func TestBlock_FailsWithoutFacade(t *testing.T) {
	t.Parallel()
	ws := startNodeServer(nil)
	defer func() {
		r := recover()
		assert.NotNil(t, r, "Not providing numbatFacade context should panic")
	}()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/block/%s", "test"), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)
}

func TestBlock_FailsWithWrongFacadeTypeConversion(t *testing.T) {
	t.Parallel()
	ws := startNodeServerWrongFacade()
	req, _ := http.NewRequest("GET", fmt.Sprintf("/block/%s", "test"), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	statusRsp := errorResponse{}
	loadResponse(resp.Body, &statusRsp)
	assert.Equal(t, resp.Code, http.StatusInternalServerError)
	assert.Equal(t, apiErrors.ErrInvalidAppContext.Error(), statusRsp.Error)
}

func TestBlock_ReturnsCorrectly(t *testing.T) {
	t.Parallel()

	testBlockHashHex := []byte("aaee")
	facade := mock.Facade{
		RetrieveShardBlockHandler: func(blockHash []byte) (info *external.ShardBlockInfo, e error) {
			blockHashConverted, _ := hex.DecodeString(string(testBlockHashHex))
			assert.Equal(t, blockHashConverted, blockHash)
			return &external.ShardBlockInfo{
				BlockHeader: external.BlockHeader{
					Nonce: 1,
				},
			}, nil
		},
	}

	ws := startNodeServer(&facade)
	req, _ := http.NewRequest("GET", fmt.Sprintf("/block/%s", testBlockHashHex), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	rb := recentBlocksResponse{}
	loadResponse(resp.Body, &rb)
	assert.Equal(t, resp.Code, http.StatusOK)
	assert.NotNil(t, rb.ShardHeader)
	assert.Equal(t, uint64(1), rb.ShardHeader.Nonce)
}

func TestBlock_KeyNotFoundShouldReturnPageNotFound(t *testing.T) {
	t.Parallel()

	testBlockHashHex := []byte("aaee")
	facade := mock.Facade{
		RetrieveShardBlockHandler: func(blockHash []byte) (info *external.ShardBlockInfo, e error) {
			return &external.ShardBlockInfo{}, errors.New("not found")
		},
	}

	ws := startNodeServer(&facade)
	req, _ := http.NewRequest("GET", fmt.Sprintf("/block/%s", testBlockHashHex), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	rb := recentBlocksResponse{}
	loadResponse(resp.Body, &rb)
	assert.Equal(t, resp.Code, http.StatusNotFound)
}

func TestBlock_KeyIsNotHexShouldReturnServerError(t *testing.T) {
	t.Parallel()

	testBlockHashHex := []byte("aae_")
	facade := mock.Facade{
		RetrieveShardBlockHandler: func(blockHash []byte) (info *external.ShardBlockInfo, e error) {
			return &external.ShardBlockInfo{}, errors.New("not found")
		},
	}

	ws := startNodeServer(&facade)
	req, _ := http.NewRequest("GET", fmt.Sprintf("/block/%s", testBlockHashHex), nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	rb := recentBlocksResponse{}
	loadResponse(resp.Body, &rb)
	assert.Equal(t, resp.Code, http.StatusInternalServerError)
}
