package goscore

import (
	"encoding/xml"
	"strconv"
	"sync"
)

// RandomForest - PMML Random Forest
type RandomForest struct {
	XMLName xml.Name
	Trees   []Node `xml:"MiningModel>Segmentation>Segment>TreeModel"`
}

// LabelScores - traverses all trees in RandomForest with features and maps result
// labels to how many trees returned those label
func (rf RandomForest) LabelScores(features map[string]interface{}) (map[string]float64, error) {
	scores := map[string]float64{}
	for _, tree := range rf.Trees {
		score, err := tree.TraverseTree(features)
		if err != nil {
			return scores, err
		}
		scoreString := strconv.FormatFloat(score, 'f', -1, 64)
		scores[scoreString]++
	}
	return scores, nil
}

// Score - traverses all trees in RandomForest with features and returns ratio of
// given label results count to all results count
func (rf RandomForest) Score(features map[string]interface{}, label string) (float64, error) {
	labelScores, err := rf.LabelScores(features)

	allCount := 0.0
	for _, value := range labelScores {
		allCount += value
	}

	return labelScores[label] / allCount, err
}

// ScoreConcurrently - same as Score but concurrent
func (rf RandomForest) ScoreConcurrently(features map[string]interface{}, label string) (float64, error) {
	labelScores, err := rf.LabelScoresConcurrently(features)

	allCount := 0.0
	for _, value := range labelScores {
		allCount += value
	}

	return labelScores[label] / allCount, err
}

type rfResult struct {
	ErrorName error
	Score     string
}

// LabelScoresConcurrently - same as LabelScores but concurrent
func (rf RandomForest) LabelScoresConcurrently(features map[string]interface{}) (map[string]float64, error) {
	messages := make(chan rfResult, len(rf.Trees))

	var wg sync.WaitGroup
	wg.Add(len(rf.Trees))

	scores := map[string]float64{}
	for _, tree := range rf.Trees {
		go func(tree Node, features map[string]interface{}) {
			treeScore, err := tree.TraverseTree(features)
			scoreString := strconv.FormatFloat(treeScore, 'f', -1, 64)
			messages <- rfResult{ErrorName: err, Score: scoreString}
			wg.Done()
		}(tree, features)
	}
	wg.Wait()

	var res rfResult
	for i := 0; i < len(rf.Trees); i++ {
		res = <-messages
		if res.ErrorName != nil {
			return map[string]float64{}, res.ErrorName
		}
		scores[res.Score]++
	}

	return scores, nil
}
