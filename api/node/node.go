// Copyright (c) 2018 The VeChainThor developers

// Distributed under the GNU Lesser General Public License v3.0 software license, see the accompanying
// file LICENSE or <https://www.gnu.org/licenses/lgpl-3.0.html>

package node

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/vechain/thor/api/utils"
	"github.com/vechain/thor/api/transactions"
	"github.com/vechain/thor/txpool"

)

type Node struct {
	nw Network
	pool *txpool.TxPool
}

func New(nw Network, pool *txpool.TxPool) *Node {
	return &Node{
		nw,
		pool,
	}
}

func (n *Node) PeersStats() []*PeerStats {
	return ConvertPeersStats(n.nw.PeersStats())
}

func (n *Node) handleNetwork(w http.ResponseWriter, req *http.Request) error {
	return utils.WriteJSON(w, n.PeersStats())
}

func (n *Node) handleGetAllTxIDsFromTxPool(w http.ResponseWriter, req *http.Request) error {
	expanded := req.URL.Query().Get("expanded")

	if expanded != "" && expanded != "false" && expanded != "true" {
		return utils.BadRequest(errors.New("expanded should be boolean"))
	}

	expandedBool := expanded == "true"

	txs := n.pool.Dump()
	if expandedBool {
		detailedTxs := make([]*transactions.Transaction, 0, len(txs))
		for _, tx := range txs {
			detailedTxs = append(detailedTxs, transactions.ConvertTransaction(tx, nil))
		}
		return utils.WriteJSON(w, detailedTxs)
	} else {
		txIDs := make([]string, 0, len(txs))
		for _, tx := range txs {
			txIDs = append(txIDs, tx.ID().String())
		}
		return utils.WriteJSON(w, txIDs)
	}
}

func (n *Node) Mount(root *mux.Router, pathPrefix string) {
	sub := root.PathPrefix(pathPrefix).Subrouter()

	sub.Path("/network/peers").Methods("Get").HandlerFunc(utils.WrapHandlerFunc(n.handleNetwork))
	sub.Path("/txpool").Methods("GET").HandlerFunc(utils.WrapHandlerFunc(n.handleGetAllTxIDsFromTxPool))
}
