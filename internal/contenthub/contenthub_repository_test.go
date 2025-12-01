package contenthub

import (
	"testing"

	"github.com/IsaacDSC/featureflag/pkg/testrepository"
)

func TestContentHubRepository(t *testing.T) {
	const contenthubPath = "contenthub_test.json"
	ts := testrepository.NewSetupRepositoryTest(contenthubPath)
	ts.Setup()
	defer ts.TearDown()
	repo := NewContentHubRepository(contenthubPath)

	// Arrange
	testCases := []struct {
		Name       string
		Contenthub Entity
		IsError    bool
	}{
		{
			Name: "Should be able to save content hub",
			Contenthub: NewEntity(
				true,
				"test",
				"test",
				"test",
				SessionsStrategies{
					{SessionID: "default", Response: "default-value"},
				},
				BalancerStrategy{
					{Weight: 50, Response: "response-a"},
					{Weight: 50, Response: "response-b"},
				},
			),
			IsError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Act
			err := repo.SaveContentHub(tc.Contenthub)

			// Assert
			if err != nil {
				t.Errorf("SaveContentHub() error = %v, wantErr %v", err, false)
			}

			results, err := repo.GetAllContentHub()
			_, ok := results[tc.Contenthub.Variable]
			if !ok {
				t.Errorf("Not found contenthub with key: %s", tc.Contenthub.Variable)
			}
		})
	}

}
