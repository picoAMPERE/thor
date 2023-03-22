// Copyright (c) 2018 The VeChainThor developers

// Distributed under the GNU Lesser General Public License v3.0 software license, see the accompanying
// file LICENSE or <https://www.gnu.org/licenses/lgpl-3.0.html>
package node_test

import (
	"encoding/json"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/vechain/thor/api/node"
	"github.com/vechain/thor/chain"
	"github.com/vechain/thor/comm"
	"github.com/vechain/thor/genesis"
	"github.com/vechain/thor/muxdb"
	"github.com/vechain/thor/state"
	"github.com/vechain/thor/thor"
	"github.com/vechain/thor/tx"
	"github.com/vechain/thor/txpool"
)

var repo *chain.Repository
var ts *httptest.Server
var txPool *txpool.TxPool

func TestNode(t *testing.T) {
	initCommServer(t)
	getTxPending(t)
	res := httpGet(t, ts.URL+"/node/network/peers")
	var peersStats map[string]string
	if err := json.Unmarshal(res, &peersStats); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 0, len(peersStats), "count should be zero")
}

func initCommServer(t *testing.T) {
	db := muxdb.NewMem()
	stater := state.NewStater(db)
	gene := genesis.NewDevnet()

	b, _, _, err := gene.Build(stater)
	if err != nil {
		t.Fatal(err)
	}
	localRepo, _ := chain.NewRepository(db, b)
	repo = localRepo
	txPool = txpool.New(repo, stater, txpool.Options{
		Limit:           10000,
		LimitPerAccount: 16,
		MaxLifetime:     10 * time.Minute,
	})
	comm := comm.New(repo, txPool)
	router := mux.NewRouter()
	node.New(comm, txPool).Mount(router, "/node")
	ts = httptest.NewServer(router)
}

func httpGet(t *testing.T, url string) []byte {
	res, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	r, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func getTxPending(t *testing.T) {
	addr := thor.BytesToAddress([]byte("to"))
	cla := tx.NewClause(&addr).WithValue(big.NewInt(10000))
	builder := new(tx.Builder).
		ChainTag(repo.ChainTag()).
		GasPriceCoef(1).
		Expiration(10).
		Gas(21000).
		Nonce(1).
		Clause(cla).
		BlockRef(tx.NewBlockRef(0))

	tx := builder.Build()

	sig, err := crypto.Sign(tx.SigningHash().Bytes(), genesis.DevAccounts()[0].PrivateKey)
	if err != nil {
		t.Fatal(err)
	}
	tx = tx.WithSignature(sig)

	if err := txPool.Add(tx); err != nil {
		t.Fatal(err)
	}

	res := httpGet(t, ts.URL+"/node/txpool")
	var poolIds []string
	if err := json.Unmarshal(res, &poolIds); err != nil {
		t.Fatal(err)
	}
	assert.Contains(t, poolIds, tx.ID().String(), "should contain new transaction id")
}
