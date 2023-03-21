package scenarios

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"math/rand"
	"testing"
)

func TestGetOrderedStepsByNotExistedTriggerId(t *testing.T) {
	service := &ScenarioService{steps: generateSteps()}
	require.Equal(t, 0, len(service.GetOrderedStepsByTriggerId(-1)), "Should be 0 because there is no trigger with id = -1")
}

func BenchmarkGetOrderedStepsByTriggerId(b *testing.B) {
	service := &ScenarioService{steps: generateSteps()}
	for i := 0; i < 10; i++ {
		input := rand.Int63n(1000)
		b.Run(fmt.Sprintf("input_%d", input), func(b *testing.B) {
			b.ReportAllocs()
			for j := 0; j < b.N; j++ {
				service.GetOrderedStepsByTriggerId(input)
			}
		})
	}
}

func generateSteps() map[int64]Steps {
	res := make(map[int64]Steps)
	for i := int64(0); i < 1000; i++ {
		max := rand.Intn(3) + 4
		for j := 0; j <= rand.Intn(3)+4; j++ {
			res[i] = append(res[i], &ScenarioStep{
				Id:          i + int64(j),
				OrderNumber: max - j,
				TriggerId:   i,
			})
		}
	}
	return res
}
