package transaction

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/numbatx/gn-numbat/api/errors"
	"github.com/numbatx/gn-numbat/data/transaction"
)

// TxService interface defines methods that can be used from `numbatFacade` context variable
type TxService interface {
	GenerateTransaction(sender string, receiver string, value *big.Int, code string) (*transaction.Transaction, error)
	SendTransaction(nonce uint64, sender string, receiver string, value *big.Int, code string, signature []byte) (*transaction.Transaction, error)
	GetTransaction(hash string) (*transaction.Transaction, error)
	GenerateAndSendBulkTransactions(string, *big.Int, uint64) error
	GenerateAndSendBulkTransactionsOneByOne(string, *big.Int, uint64) error
}

// TxRequest represents the structure on which user input for generating a new transaction will validate against
type TxRequest struct {
	Sender   string   `form:"sender" json:"sender"`
	Receiver string   `form:"receiver" json:"receiver"`
	Value    *big.Int `form:"value" json:"value"`
	Data     string   `form:"data" json:"data"`
	//SecretKey string `form:"sk" json:"sk" binding:"skValidator"`
}

// MultipleTxRequest represents the structure on which user input for generating a bulk of transactions will validate against
type MultipleTxRequest struct {
	Receiver string   `form:"receiver" json:"receiver"`
	Value    *big.Int `form:"value" json:"value"`
	TxCount  int      `form:"txCount" json:"txCount"`
}

// SendTxRequest represents the structure that maps and validates user input for publishing a new transaction
type SendTxRequest struct {
	Sender    string   `form:"sender" json:"sender"`
	Receiver  string   `form:"receiver" json:"receiver"`
	Value     *big.Int `form:"value" json:"value"`
	Data      string   `form:"data" json:"data"`
	Nonce     uint64   `form:"nonce" json:"nonce"`
	GasPrice  *big.Int `form:"gasPrice" json:"gasPrice"`
	GasLimit  *big.Int `form:"gasLimit" json:"gasLimit"`
	Signature string   `form:"signature" json:"signature"`
	Challenge string   `form:"challenge" json:"challenge"`
}

// TxResponse represents the structure on which the response will be validated against
type TxResponse struct {
	SendTxRequest
	ShardID     uint32 `json:"shardId"`
	Hash        string `json:"hash"`
	BlockNumber uint64 `json:"blockNumber"`
	BlockHash   string `json:"blockHash"`
	Timestamp   uint64 `json:"timestamp"`
}

// Routes defines transaction related routes
func Routes(router *gin.RouterGroup) {
	router.POST("/generate", GenerateTransaction)
	router.POST("/generate-and-send-multiple", GenerateAndSendBulkTransactions)
	router.POST("/generate-and-send-multiple-one-by-one", GenerateAndSendBulkTransactionsOneByOne)
	router.POST("/send", SendTransaction)
	router.GET("/:txhash", GetTransaction)
}

// RoutesForTransactionsLists defines routes related to lists of transactions. Used separately so
// it will not conflict with the wildcard for transaction details route
func RoutesForTransactionsLists(router *gin.RouterGroup) {
	router.GET("/recent", RecentTransactions)
}

// GenerateTransaction generates a new transaction given a sender, receiver, value and data
func GenerateTransaction(c *gin.Context) {
	ef, ok := c.MustGet("numbatFacade").(TxService)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errors.ErrInvalidAppContext.Error()})
		return
	}

	var gtx = TxRequest{}
	err := c.ShouldBindJSON(&gtx)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), err.Error())})
		return
	}

	tx, err := ef.GenerateTransaction(gtx.Sender, gtx.Receiver, gtx.Value, gtx.Data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%s: %s", errors.ErrTxGenerationFailed.Error(), err.Error())})
		return
	}

	c.JSON(http.StatusOK, gin.H{"transaction": txResponseFromTransaction(tx)})
}

// SendTransaction will receive a transaction from the client and propagate it for processing
func SendTransaction(c *gin.Context) {
	ef, ok := c.MustGet("numbatFacade").(TxService)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errors.ErrInvalidAppContext.Error()})
		return
	}

	var gtx = SendTxRequest{}
	err := c.ShouldBindJSON(&gtx)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), err.Error())})
		return
	}

	signature, err := hex.DecodeString(gtx.Signature)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%s: %s", errors.ErrInvalidSignatureHex.Error(), err.Error())})
		return
	}

	tx, err := ef.SendTransaction(gtx.Nonce, gtx.Sender, gtx.Receiver, gtx.Value, gtx.Data, signature)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%s: %s", errors.ErrTxGenerationFailed.Error(), err.Error())})
		return
	}

	c.JSON(http.StatusOK, gin.H{"transaction": txResponseFromTransaction(tx)})
}

// GenerateAndSendBulkTransactions generates multipleTransactions
func GenerateAndSendBulkTransactions(c *gin.Context) {
	ef, ok := c.MustGet("numbatFacade").(TxService)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errors.ErrInvalidAppContext.Error()})
		return
	}

	var gtx = MultipleTxRequest{}
	err := c.ShouldBindJSON(&gtx)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), err.Error())})
		return
	}

	err = ef.GenerateAndSendBulkTransactions(gtx.Receiver, gtx.Value, uint64(gtx.TxCount))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%s: %s", errors.ErrMultipleTxGenerationFailed.Error(), err.Error())})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("%d", gtx.TxCount)})
}

// GenerateAndSendBulkTransactionsOneByOne generates multipleTransactions in a one-by-one fashion
func GenerateAndSendBulkTransactionsOneByOne(c *gin.Context) {
	ef, ok := c.MustGet("numbatFacade").(TxService)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errors.ErrInvalidAppContext.Error()})
		return
	}

	var gtx = MultipleTxRequest{}
	err := c.ShouldBindJSON(&gtx)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), err.Error())})
		return
	}

	err = ef.GenerateAndSendBulkTransactionsOneByOne(gtx.Receiver, gtx.Value, uint64(gtx.TxCount))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("%s: %s", errors.ErrMultipleTxGenerationFailed.Error(), err.Error())})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("%d", gtx.TxCount)})
}

// GetTransaction returns transaction details for a given txhash
func GetTransaction(c *gin.Context) {

	ef, ok := c.MustGet("numbatFacade").(TxService)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errors.ErrInvalidAppContext.Error()})
		return
	}

	txhash := c.Param("txhash")
	if txhash == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("%s: %s", errors.ErrValidation.Error(), errors.ErrValidationEmptyTxHash.Error())})
		return
	}

	tx, err := ef.GetTransaction(txhash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errors.ErrGetTransaction.Error()})
		return
	}

	if tx == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": errors.ErrTxNotFound.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"transaction": txResponseFromTransaction(tx)})
}

// RecentTransactions returns the list of latest transactions from all shards
func RecentTransactions(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"transactions": buildDummyRecentTransactions()})
}

func buildDummyRecentTransactions() []TxResponse {
	txs := make([]TxResponse, 0)
	for i := 0; i < 10; i++ {
		txs = append(txs, TxResponse{
			SendTxRequest{
				Sender:    "0x000000",
				Receiver:  "0x11111",
				Value:     big.NewInt(10),
				Data:      "",
				Nonce:     1,
				GasPrice:  big.NewInt(10),
				GasLimit:  big.NewInt(10),
				Signature: "0x12314212313",
			},
			1,
			"0x3213894328492",
			10,
			"0x000000000",
			1558361492,
		})
	}
	return txs
}

func txResponseFromTransaction(tx *transaction.Transaction) TxResponse {
	response := TxResponse{}
	response.Nonce = tx.Nonce
	response.Sender = hex.EncodeToString(tx.SndAddr)
	response.Receiver = hex.EncodeToString(tx.RcvAddr)
	response.Data = string(tx.Data)
	response.Signature = hex.EncodeToString(tx.Signature)
	response.Challenge = string(tx.Challenge)
	response.Value = tx.Value
	response.GasLimit = big.NewInt(int64(tx.GasLimit))
	response.GasPrice = big.NewInt(int64(tx.GasPrice))

	return response
}
