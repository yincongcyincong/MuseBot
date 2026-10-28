package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReplaceAndGetMemories(t *testing.T) {
	userId := "memory-test-user"

	err := ReplaceMemories(context.Background(), userId, []string{"likes go", "works on musebot"})
	assert.NoError(t, err)

	memories := GetMemories(userId)
	assert.Len(t, memories, 2)
	assert.Equal(t, "likes go", memories[0].Content)
	assert.Equal(t, "works on musebot", memories[1].Content)

	// replace drops old rows and keeps order of the new set
	err = ReplaceMemories(context.Background(), userId, []string{"prefers zh", "", "runs on linux"})
	assert.NoError(t, err)

	memories = GetMemories(userId)
	assert.Len(t, memories, 2)
	assert.Equal(t, "prefers zh", memories[0].Content)
	assert.Equal(t, "runs on linux", memories[1].Content)
}
