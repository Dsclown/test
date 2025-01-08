package main

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"
)

type RankInfo struct {
	PlayerId  string
	Score     int
	Timestamp time.Time

	RankNum int
}

type Leaderboard struct {
	rankMap  sync.Map
	quickPic []RankInfo
}

var (
	instance *Leaderboard
	once     sync.Once
)

// 单例保证数据唯一
func GetInstance() *Leaderboard {
	once.Do(func() {
		instance = &Leaderboard{
			rankMap:  sync.Map{},
			quickPic: make([]RankInfo, 0),
		}
	})
	return instance
}

// 定时更新快照，查询从快照中查询
func UpdateQuickPic(lb *Leaderboard) {
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for range ticker.C {
			lb.quickPic = make([]RankInfo, 0)
			lb.rankMap.Range(func(key, value interface{}) bool {
				lb.quickPic = append(lb.quickPic, value.(RankInfo))
				return true
			})
			print("update quick pic done")
		}
	}()
}

// 插入只更新map
func (lb *Leaderboard) updateScore(playerId string, score int, timestamp time.Time) {

	lb.rankMap.Store(playerId, RankInfo{PlayerId: playerId, Score: score, Timestamp: timestamp})
	fmt.Printf("\nupdate score done: %s %d %d\n", playerId, score, timestamp.UnixNano())
}

// 查询时，从快照中查询（额外取玩家自身最新分数）
func (lb *Leaderboard) getTopN(playerId string, n int) []RankInfo {

	tempRankInfo := lb.quickPic
	curPlayerInfo, ok := lb.rankMap.Load(playerId)
	if ok {
		flag := false
		for index, rankInfo := range tempRankInfo {
			if rankInfo.PlayerId == playerId {
				tempRankInfo[index] = curPlayerInfo.(RankInfo)
				flag = true
				break
			}
		}
		if !flag {
			tempRankInfo = append(tempRankInfo, curPlayerInfo.(RankInfo))
		}
	}

	sort.Slice(tempRankInfo, func(i, j int) bool {
		if tempRankInfo[i].Score == tempRankInfo[j].Score {
			return tempRankInfo[i].Timestamp.Before(tempRankInfo[j].Timestamp)
		}
		return tempRankInfo[i].Score > tempRankInfo[j].Score
	})

	for rankNum := range tempRankInfo {
		tempRankInfo[rankNum].RankNum = rankNum + 1
	}

	if n > len(tempRankInfo) {
		n = len(tempRankInfo)
	}

	return tempRankInfo[:n]
}

func (lb *Leaderboard) getPlayerRankRange(playerId string, n int) []RankInfo {

	tempRankInfo := lb.quickPic
	curPlayerInfo, ok := lb.rankMap.Load(playerId)
	if ok {
		flag := false
		for index, rankInfo := range tempRankInfo {
			if rankInfo.PlayerId == playerId {
				tempRankInfo[index] = curPlayerInfo.(RankInfo)
				flag = true
				break
			}
		}
		if !flag {
			tempRankInfo = append(tempRankInfo, curPlayerInfo.(RankInfo))
		}
	}

	sort.Slice(tempRankInfo, func(i, j int) bool {
		if tempRankInfo[i].Score == tempRankInfo[j].Score {
			return tempRankInfo[i].Timestamp.Before(tempRankInfo[j].Timestamp)
		}
		return tempRankInfo[i].Score > tempRankInfo[j].Score
	})

	for rankNum := range tempRankInfo {
		tempRankInfo[rankNum].RankNum = rankNum + 1
	}

	var targetIndex int
	for i, player := range tempRankInfo {
		if player.PlayerId == playerId {
			targetIndex = i
			break
		}
	}

	temp := n / 2
	start := targetIndex - temp
	if start < 0 {
		start = 0
		temp = targetIndex - start
	}
	end := targetIndex + (n - temp)
	if end > len(tempRankInfo) {
		end = len(tempRankInfo)
	}

	return tempRankInfo[start:end]
}

func (lb *Leaderboard) getPlayerRank(playerId string) int {

	tempRankInfo := lb.quickPic
	curPlayerInfo, ok := lb.rankMap.Load(playerId)
	if ok {
		flag := false
		for index, rankInfo := range tempRankInfo {
			if rankInfo.PlayerId == playerId {
				tempRankInfo[index] = curPlayerInfo.(RankInfo)
				flag = true
				break
			}
		}
		if !flag {
			tempRankInfo = append(tempRankInfo, curPlayerInfo.(RankInfo))
		}
	}

	sort.Slice(tempRankInfo, func(i, j int) bool {
		if tempRankInfo[i].Score == tempRankInfo[j].Score {
			return tempRankInfo[i].Timestamp.Before(tempRankInfo[j].Timestamp)
		}
		return tempRankInfo[i].Score > tempRankInfo[j].Score
	})

	for i, player := range tempRankInfo {
		if player.PlayerId == playerId {
			return i + 1
		}
	}

	return -1
}

// Reset clears the leaderboard
func (lb *Leaderboard) Reset() {

	lb.rankMap = sync.Map{}
	lb.quickPic = make([]RankInfo, 0)
}

func main() {
	lb := GetInstance()
	UpdateQuickPic(lb)

	/*

		lb.updateScore("Alice", 50, time.Now())
		lb.updateScore("Bob", 30, time.Now())
		lb.updateScore("Alice", 10, time.Now())
		lb.updateScore("Charlie", 40, time.Now())
		lb.updateScore("Dave", 60, time.Now())
		lb.updateScore("Eve", 10, time.Now())

		fmt.Println("Top 3 Players:")
		topPlayers := lb.getTopN(3)
		for _, rankInfo := range topPlayers {
			fmt.Printf("RankNum: %d, Name: %s, Score: %d, UpdateTime: %d\n", rankInfo.RankNum, rankInfo.PlayerId, rankInfo.Score, rankInfo.Timestamp.UnixNano())
		}

		fmt.Println("\nTop 10 Players:")
		topPlayers = lb.getTopN(10)
		for _, rankInfo := range topPlayers {
			fmt.Printf("RankNum: %d, Name: %s, Score: %d, UpdateTime: %d\n", rankInfo.RankNum, rankInfo.PlayerId, rankInfo.Score, rankInfo.Timestamp.UnixNano())
		}

		fmt.Println("\nSurrounding 3 Players of Charlie:")
		surroundingPlayers := lb.getPlayerRankRange("Charlie", 3)
		for _, rankInfo := range surroundingPlayers {
			fmt.Printf("RankNum: %d, Name: %s, Score: %d, UpdateTime: %d\n", rankInfo.RankNum, rankInfo.PlayerId, rankInfo.Score, rankInfo.Timestamp.UnixNano())
		}

		fmt.Println("\nRank of Charlie:")
		rank := lb.getPlayerRank("Charlie")
		fmt.Printf("Charlie is ranked %d\n", rank)

	*/

	// 模拟多用户并发
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		playerId := "player" + fmt.Sprintf("%02d", i)
		wg.Add(1)
		go func(playerId string) {

			defer wg.Done()

			// 模拟玩家分数更新（不定时）
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(10000)))
			lb.updateScore(playerId, rand.Intn(100), time.Now())

			var output string
			output += fmt.Sprintf("Top 3 Players:\n")
			topPlayers := lb.getTopN(playerId, 3)
			for _, rankInfo := range topPlayers {
				output += fmt.Sprintf("RankNum: %d, Name: %s, Score: %d, UpdateTime: %d\n", rankInfo.RankNum, rankInfo.PlayerId, rankInfo.Score, rankInfo.Timestamp.UnixMilli())
			}
			output += fmt.Sprintf("\nSurrounding 3 Players of %s:\n", playerId)
			surroundingPlayers := lb.getPlayerRankRange(playerId, 3)
			for _, rankInfo := range surroundingPlayers {
				output += fmt.Sprintf("RankNum: %d, Name: %s, Score: %d, UpdateTime: %d\n", rankInfo.RankNum, rankInfo.PlayerId, rankInfo.Score, rankInfo.Timestamp.UnixMilli())
			}
			rank := lb.getPlayerRank(playerId)
			output += fmt.Sprintf("\n%s is ranked %d\n", playerId, rank)

			fmt.Println(output)
		}(playerId)
	}

	wg.Wait()

	fmt.Println("\nTop 10 Players:")
	topPlayers := lb.getTopN("", 10)
	for _, rankInfo := range topPlayers {
		fmt.Printf("RankNum: %d, Name: %s, Score: %d, UpdateTime: %d\n", rankInfo.RankNum, rankInfo.PlayerId, rankInfo.Score, rankInfo.Timestamp.UnixMilli())
	}

	lb.Reset()
}
