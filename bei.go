package tabula

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"strconv"

	"code.rocket9labs.com/tslocum/bei"
)

type BEIServer struct {
}

func NewBEIServer() *BEIServer {
	return &BEIServer{}
}

func (s *BEIServer) handleConnection(conn net.Conn) {
	analysis := make([]*Analysis, 0, AnalysisBufferSize)
	var beiCommand bool
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		if scanner.Err() != nil {
			log.Printf("error: failed to read from client: %s", scanner.Err())
			conn.Close()
			return
		}
		if !beiCommand && !bytes.Equal(scanner.Bytes(), []byte("bei")) {
			log.Printf("error: failed to read from client: failed to receive bei command")
			conn.Close()
			return
		}
		switch {
		case bytes.Equal(scanner.Bytes(), []byte("bei")):
			buf, err := bei.EncodeEvent(&bei.EventOkBEI{
				Version: 1,
				ID: map[string]string{
					"name": "tabula",
				},
			})
			if err != nil {
				log.Fatalf("error: failed to encode event: %s", err)
			}
			conn.Write(buf)
			conn.Write([]byte("\n"))
			beiCommand = true
		case bytes.HasPrefix(scanner.Bytes(), []byte("move ")):
			b, err := parseState(scanner.Bytes()[5:])
			if err != nil {
				log.Println(err)
				conn.Close()
				return
			}

			available, _ := b.Available(1)
			b.Analyze(available, &analysis)
			if len(analysis) == 0 {
				log.Printf("error: failed to read from client: zero moves returned for analysis")
				conn.Close()
				return
			}

			move := &bei.Move{}
			for _, m := range analysis[0].Moves {
				if m[0] == 0 && m[1] == 0 {
					break
				}
				move.Play = append(move.Play, &bei.Play{From: int(m[0]), To: int(m[1])})
			}
			buf, err := bei.EncodeEvent(&bei.EventOkMove{
				Moves: []*bei.Move{move},
			})
			if err != nil {
				log.Fatalf("error: failed to encode event: %s", err)
			}
			conn.Write(buf)
			conn.Write([]byte("\n"))
		case bytes.HasPrefix(scanner.Bytes(), []byte("choose ")):
			b, err := parseState(scanner.Bytes()[7:])
			if err != nil {
				log.Println(err)
				conn.Close()
				return
			}

			if b[SpaceAcey] != 1 {
				log.Println("error: failed to choose roll: state does not represent acey-deucey game")
				conn.Close()
				return
			}

			roll := b.ChooseDoubles(&analysis)
			if roll < 1 || roll > 6 {
				log.Printf("error: failed to read from client: invalid roll: %d", roll)
				conn.Close()
				return
			}

			buf, err := bei.EncodeEvent(&bei.EventOkChoose{
				Rolls: []*bei.ChooseRoll{
					{
						Roll: roll,
					},
				},
			})
			if err != nil {
				log.Fatalf("error: failed to encode event: %s", err)
			}
			conn.Write(buf)
			conn.Write([]byte("\n"))
		default:
			log.Printf("error: received unexpected command from client: %s", scanner.Bytes())
			conn.Close()
			return
		}
	}
	if scanner.Err() != nil {
		log.Printf("error: failed to read from client: %s", scanner.Err())
		conn.Close()
		return
	}
}

func (s *BEIServer) Listen(address string) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen on %s: %s", address, err)
	}
	log.Printf("Listening for connections on %s...", address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("failed to listen on %s: %s", address, err)
		}

		go s.handleConnection(conn)
	}
}

func parseState(buf []byte) (Board, error) {
	var stateInts []int
	for _, v := range bytes.Split(buf, []byte(",")) {
		i, err := strconv.Atoi(string(v))
		if err != nil {
			return Board{}, fmt.Errorf("error: failed to read from client: failed to decode state: %s", err)
		}
		stateInts = append(stateInts, i)
	}
	state, err := bei.DecodeState(stateInts)
	if err != nil {
		return Board{}, fmt.Errorf("error: failed to read from client: failed to decode state: %s", err)
	}
	b := Board{}
	for i, v := range state.Board {
		b[i] = int8(v)
	}
	b[SpaceRoll1] = int8(state.Roll1)
	b[SpaceRoll2] = int8(state.Roll2)
	if state.Roll1 == state.Roll2 {
		b[SpaceRoll3], b[SpaceRoll4] = int8(state.Roll1), int8(state.Roll2)
	}
	b[SpaceEnteredPlayer] = 1
	b[SpaceEnteredOpponent] = 1
	if state.Acey {
		b[SpaceAcey] = 1
		if !state.Entered1 {
			b[SpaceEnteredPlayer] = 0
		}
		if !state.Entered2 {
			b[SpaceEnteredOpponent] = 0
		}
	}
	return b, nil
}
