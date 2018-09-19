package lobby

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLobby_PickCaptains(t *testing.T) {
	l := NewLobby()
	l.Players = append(l.Players, NewPlayer("a", "a"))
	l.Players = append(l.Players, NewPlayer("b", "b"))
	l.Limit = 2
	err := l.PickCaptains()
	assert.NoError(t, err)

	rc := l.GetCaptain(RED)
	bc := l.GetCaptain(BLU)
	assert.NotEqual(t, rc.ID, bc.ID)
}
