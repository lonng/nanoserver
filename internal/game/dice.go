package game

import "math/rand"

type dice struct {
	dice1 int
	dice2 int
}

func newDice() *dice {
	return &dice{}
}

func (d *dice) random() {
	d.dice1, d.dice2 = rand.Intn(6)+1, rand.Intn(6)+1
}
