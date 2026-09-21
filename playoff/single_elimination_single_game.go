// Copyright 2026 Team 254. All Rights Reserved.
//
// Defines the tournament structure for a single-elimination bracket whose rounds are decided by a single game,
// except for the final, which remains best-of-three.

package playoff

// Creates a single-elimination bracket with the same structure as newSingleEliminationBracket, but with every
// matchup before the final requiring only one win to advance. The second and third matches of those matchups start
// hidden; matchup.update will only reveal the second match if the first is tied, and the third only if the second is
// also tied.
func newSingleEliminationSingleGameBracket(numAlliances int) (*Matchup, []breakSpec, error) {
	final, breakSpecs, err := newSingleEliminationBracket(numAlliances)
	if err != nil {
		return nil, nil, err
	}

	err = final.traverse(
		func(matchGroup MatchGroup) error {
			matchup, ok := matchGroup.(*Matchup)
			if !ok || matchup.isFinal() {
				// Leave the final as a best-of-three matchup.
				return nil
			}
			matchup.NumWinsToAdvance = 1
			for _, match := range matchup.matchSpecs[1:] {
				match.isHidden = true
			}
			return nil
		},
	)
	if err != nil {
		return nil, nil, err
	}

	return final, breakSpecs, nil
}
