package scenarios

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"math/rand"
	"sort"
	"sync"
	"testing"
)

func TestGetOrderedStepsByNotExistedTriggerId(t *testing.T) {
	service := &ScenarioService{steps: generateSteps(1000)}
	require.Equal(t, 0, len(service.GetOrderedStepsByTriggerId(-1)), "Should be 0 because there is no trigger with id = -1")
}

func BenchmarkGetOrderedStepsByTriggerId(b *testing.B) {
	service := &ScenarioService{steps: generateSteps(1000)}
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

func generateSteps(triggersCount int64) map[int64]Steps {
	res := make(map[int64]Steps)
	for i := int64(0); i < triggersCount; i++ {
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

func TestGetOrderedStepsByTriggerIdSort(t *testing.T) {
	service := &ScenarioService{steps: generateSteps(1000)}
	service.steps[0].Print()
	service.steps[0][0], service.steps[0][1] = service.steps[0][1], service.steps[0][0]
	service.steps[0].Print()
	var wg sync.WaitGroup
	wg.Add(100000)
	for i := 0; i < 100000; i++ {
		go func() {
			_ = service.GetOrderedStepsByTriggerId(0)
			wg.Done()
		}()
	}
	wg.Wait()
	service.steps[0].Print()
}

func TestSortSliceConcurrently(t *testing.T) {
	steps := []int{2, 3, 5, 1, 3}
	var wg sync.WaitGroup
	wg.Add(100000)
	for i := 0; i < 100000; i++ {
		go func() {
			sort.Slice(steps, func(i, j int) bool {
				return steps[i] < steps[j]
			})
			wg.Done()
		}()
	}
	wg.Wait()
	fmt.Printf("%v", steps)
}

func (steps Steps) Print() {
	for i := range steps {
		fmt.Printf("%d ", steps[i].OrderNumber)
	}
	fmt.Println()
}
