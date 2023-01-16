import (
    "fmt"
)

type coords struct {
    x, y int
}

type elementset map[int8]bool

func (es *elementset) get() int8 {
    for el := range *es {
        return el
    }
    return 0
}


func (es *elementset) add(el int8) {
    (*es)[el] = true
}


var btoi map[byte]int8 = map[byte]int8{
    byte('.'): -1,
    byte('1'): 0,
    byte('2'): 1,
    byte('3'): 2,
    byte('4'): 3,
    byte('5'): 4,
    byte('6'): 5,
    byte('7'): 6,
    byte('8'): 7,
    byte('9'): 8,
}


var itob map[int8]byte = map[int8]byte{
    -1: byte('.'),
    0: byte('1'),
    1: byte('2'),
    2: byte('3'),
    3: byte('4'),
    4: byte('5'),
    5: byte('6'),
    6: byte('7'),
    7: byte('8'),
    8: byte('9'),
}


func coordsToSubbox(i, j int) int {
    ip, jp := i / 3, j / 3
    return ip * 3 + jp
}


func copyBoard(board [][]byte) [][]byte {

    var new_board [][]byte

    new_board = make([][]byte, 9)

    for i := 0; i < 9; i++ {
        new_board[i] = make([]byte, 9)
        for j := 0; j < 9; j++ {
            new_board[i][j] = board[i][j]
        }
    }

    return new_board

}


func copyCandidates(candidates [][]elementset) [][]elementset {

    var new_candidates [][]elementset

    new_candidates = make([][]elementset, 9)

    for i := 0; i < 9; i++ {
        new_candidates[i] = make([]elementset, 9)
        for j := 0; j < 9; j++ {
            new_candidates[i][j] = make(elementset)
            for el := range candidates[i][j] {
                new_candidates[i][j][el] = true
            }
        }
    }

    return new_candidates

}


func setTiles(single_candidates map[coords]int8, board [][]byte, candidates [][]elementset) int {
    // this function takes a map of tile coordinates that have a single candidate, and updates the board
    // and candidate map so that the tiles are considered solved.
    // also updates tiles that become single-candidate after the above is applied.

    var exists bool
    var x, y int
    var el int8
    var num_solved int

    num_solved = 0

    for len(single_candidates) > 0 {

        // pop the first single candidate
        for sc, candidate := range single_candidates {
            //fmt.Println("processing:", sc, candidate+1)
            x = sc.x
            y = sc.y
            el = candidate
            break
        }
        delete(single_candidates, coords{x: x, y: y})

        // set the board
        board[x][y] = itob[el]
        num_solved++

        // unset candidates
        candidates[x][y] = make(elementset)

        // delete the element from the candidate list of each tile
        // in the row and column
        for k := 0; k < 9; k++ {

            _, exists = candidates[x][k][el]
            if exists {
                delete(candidates[x][k], el)
                if len(candidates[x][k]) == 1 {
                    //fmt.Println("adding candidate in setTiles():", x, k, el+1)
                    single_candidates[coords{x: x, y: k}] = candidates[x][k].get()
                }
            }

            _, exists = candidates[k][y][el]
            if exists {
                delete(candidates[k][y], el)
                if len(candidates[k][y]) == 1 {
                    //fmt.Println("adding candidate in setTiles():", k, y, el+1)
                    single_candidates[coords{x: k, y: y}] = candidates[k][y].get()
                }
            }

        }

        // delete the element from the candidate list of each tile
        // in the subbox
        x_subbox, y_subbox := x / 3, y / 3

        for i := x_subbox * 3; i < (x_subbox + 1) * 3; i++ {
            for j := y_subbox * 3; j < (y_subbox + 1) * 3; j++ {
                _, exists = candidates[i][j][el]
                if exists {
                    delete(candidates[i][j], el)
                    if len(candidates[i][j]) == 1 {
                        //fmt.Println("adding candidate in setTiles():", i, j, el+1)
                        single_candidates[coords{x: i, y: j}] = candidates[i][j].get()
                    }
                }
            }
        }

    }

    return num_solved

}


func lowHangingFruit(board [][]byte, candidates [][]elementset) (int, [][]byte) {

    // this method looks for tiles with exactly one candidate, and solves those tiles.

    var res int
    var count int
    var single_candidates map[coords]int8

    // a set of coordinates that have one candidate
    single_candidates = make(map[coords]int8)

    for i := 0; i < 9; i++ {
        for j := 0; j < 9; j++ {
            // look for tiles with 1 candidates (which can be set)
            count = len(candidates[i][j])
            if count == 0 {
                // if there are 0 candidates but the board tile is still empty,
                // we've done something wrong. bail
                if board[i][j] == byte('.') {
                    return 0, nil
                }
                // otherwise, it's a solved tile, continue
                continue
            } else if count == 1 {
                // we found a tile with 1 candidate: set the tile and update num_solved
                single_candidates[coords{x: i, y: j}] = candidates[i][j].get()
            }
        }
    }

    if len(single_candidates) > 0 {
        res = setTiles(single_candidates, board, candidates)
    }

    if res == -1 {
        fmt.Println("Got invalid result in lowHangingFruit()")
        return 0, nil
    }

    return res, board

}


func graspAtStraws(board [][]byte, candidates [][]elementset) (int, [][]byte) {

    // this function checks each row, column and sub-box if there are any elements that could only
    // fit on a specific tile

    var res int
    var subbox int
    var rows, cols, subboxes [][]*coords
    var single_candidates map[coords]int8
    var prev, curr *coords
    var exists, undo_row, undo_col, undo_subbox bool

    // rows is a 2D array of each row and each element to the number of tiles that have {el} as a candidate
    rows = make([][]*coords, 9)
    cols = make([][]*coords, 9)
    subboxes = make([][]*coords, 9)

    // see definition of {single_candidates} in other functions
    single_candidates = make(map[coords]int8)

    for i := 0; i < 9; i++ {
        rows[i] = make([]*coords, 9)
        cols[i] = make([]*coords, 9)
        subboxes[i] = make([]*coords, 9)
    }

    for i := 0; i < 9; i++ {
        for j := 0; j < 9; j++ {

            if board[i][j] == byte('.') && len(candidates[i][j]) == 0 {
                // if there are 0 candidates but the board tile is still empty,
                // we've done something wrong. bail
                return 0, nil
            }

            subbox = coordsToSubbox(i, j)
            curr = &coords{x: i, y: j}
            for el := range candidates[i][j] {

                undo_row, undo_col, undo_subbox = false, false, false

                // it's very important that the current coordinates {curr} are only
                // added to {single_candidates} if {el} is a confirmed solution for {curr}
                // it's a confirmed solution if the tile at {curr} is the only possible
                // home for {el} in the row, column, or sub-box
                if rows[i][el] == nil {
                    rows[i][el] = curr
                    single_candidates[*curr] = el
                } else {
                    undo_row = true
                }
                if cols[j][el] == nil {
                    cols[j][el] = curr
                    single_candidates[*curr] = el
                } else {
                    undo_col = true
                }
                if subboxes[subbox][el] == nil {
                    subboxes[subbox][el] = curr
                    single_candidates[*curr] = el
                } else {
                    undo_subbox = true
                }

                // if, however, we find mutltiple possible homes for {el} in
                // the row, column and sub-box, we need to keep track of the previous entries
                // we've added to {single_candidates} so we can get rid of them
                if undo_row && undo_col && undo_subbox {
                    prev = rows[i][el]
                    _, exists = single_candidates[*prev]
                    if exists {
                        delete(single_candidates, *prev)
                    }
                    prev = cols[j][el]
                    _, exists = single_candidates[*prev]
                    if exists {
                        delete(single_candidates, *prev)
                    }
                    prev = subboxes[subbox][el]
                    _, exists = single_candidates[*prev]
                    if exists {
                        delete(single_candidates, *prev)
                    }
                }

            }
        }
    }

    if len(single_candidates) > 0 {
        res = setTiles(single_candidates, board, candidates)
    }

    if res == -1 {
        fmt.Println("Got invalid result in graspAtStraws()")
        return 0, nil
    }

    return res, board

}


func shotInTheDark(board [][]byte, candidates [][]elementset) (int, [][]byte) {

    // last ditch method
    // choose a tile with a minimum number of candidates, "solve" it with
    // one of the candidates at random, and see how far we get

    var candidates_by_count [][]coords
    var imaginary_board [][]byte
    var imaginary_candidates [][]elementset
    var single_candidates map[coords]int8
    var baseline, num_solved, res int

    // initialize our stuff
    candidates_by_count = make([][]coords, 9)
    for i := 0; i < 9; i++ {
        candidates_by_count[i] = make([]coords, 0)
    }

    // map tiles by the number of candidates they have
    for i := 0; i < 9; i++ {
        for j := 0; j < 9; j++ {
            if board[i][j] != byte('.') {
                baseline++
            }
            count := len(candidates[i][j])
            if count >= 2 {
                candidates_by_count[count] = append(candidates_by_count[count], coords{x: i, y: j})
            }
        }
    }

    // iterate through {candidates_by_count}, starting with the tiles that have 2 candidates
    for i := 2; i < 9; i++ {
        // iterate through the tiles with the given number of candidates
        for _, c := range candidates_by_count[i] {
            // iterate through candidates on the given tile
            for el := range candidates[c.x][c.y] {
                num_solved = baseline
                // initialize {single_candidates} which is used to make our imaginary move
                single_candidates = make(map[coords]int8)
                single_candidates[c] = el
                // initialize our imaginary board and imaginary candidates
                imaginary_board = copyBoard(board)
                imaginary_candidates = copyCandidates(candidates)

                res = setTiles(single_candidates, imaginary_board, imaginary_candidates)
                num_solved += res
                for {
                    res, imaginary_board = lowHangingFruit(imaginary_board, imaginary_candidates)
                    if imaginary_board == nil {
                        num_solved = -1
                        break
                    }
                    if res > 0 {
                        num_solved += res
                        if num_solved >= 81 {
                            return num_solved - baseline, imaginary_board
                        }
                        continue
                    }

                    res, imaginary_board = graspAtStraws(imaginary_board, imaginary_candidates)
                    if imaginary_board == nil {
                        num_solved = -1
                        break
                    }
                    if !validate(imaginary_board) {
                        break
                    }
                    if res > 0 {
                        num_solved += res
                        if num_solved >= 81 {
                            return num_solved - baseline, imaginary_board
                        }
                        continue
                    }

                    res, imaginary_board = shotInTheDark(imaginary_board, imaginary_candidates)
                    if imaginary_board == nil {
                        num_solved = -1
                        break
                    }
                    if !validate(imaginary_board) {
                        break
                    }
                    if res > 0 {
                        num_solved += res
                        if num_solved >= 81 {
                            return num_solved - baseline, imaginary_board
                        }
                        continue
                    }

                    break

                }
            }
        }
    }

    return 0, nil

}


func validate(board [][]byte) bool {

    var rows, cols, subboxes [][]bool

    // initialize our stuff
    rows = make([][]bool, 9)
    cols = make([][]bool, 9)
    subboxes = make([][]bool, 9)
    for i := 0; i < 9; i++ {
        rows[i] = make([]bool, 9)
        cols[i] = make([]bool, 9)
        subboxes[i] = make([]bool, 9)
    }

    // fill rows, cols, and subboxes to show which elements exist
    for i := 0; i < 9; i++ {
        for j := 0; j < 9; j++ {

            // if it's empty, nothing to do
            if board[i][j] == byte('.') {
                continue
            }

            // convert our "byte" to an int8
            el := btoi[board[i][j]]

            // if any of these is already true, our board is invalid
            if rows[i][el] || cols[j][el] || subboxes[coordsToSubbox(i, j)][el] {
                return false
            }

            rows[i][el] = true
            cols[j][el] = true
            subboxes[coordsToSubbox(i, j)][el] = true
        }
    }

    return true

}


func solveSudoku(board [][]byte)  {

    var rows, cols, subboxes [][]bool
    var el int8
    var res, num_solved int
    var candidates [][]elementset
    var res_board [][]byte

    // initialize our stuff
    // rows is a temporary var for keeping track of which elements are missing from each row.
    rows = make([][]bool, 9)
    // cols keeps track of which elements are missing from each column.
    cols = make([][]bool, 9)
    // subboxes keeps track of which elements are missing from each sub-box.
    subboxes = make([][]bool, 9)
    // candidates is a 2D array of sets of the candidates that are available for each tile.
    candidates = make([][]elementset, 9)
    for i := 0; i < 9; i++ {
        rows[i] = make([]bool, 9)
        cols[i] = make([]bool, 9)
        subboxes[i] = make([]bool, 9)
        candidates[i] = make([]elementset, 9)
    }

    // fill rows, cols, and subboxes to show which elements exist
    for i := 0; i < 9; i++ {
        for j := 0; j < 9; j++ {

            // if it's empty, nothing to do
            if board[i][j] == byte('.') {
                continue
            }

            // convert our "byte" to an int8
            el = btoi[board[i][j]]

            // rows, cols, and subboxes keep track of which elements exist
            rows[i][el] = true
            cols[j][el] = true
            subboxes[coordsToSubbox(i, j)][el] = true
            num_solved++
        }
    }

    // populate candidates
    for i := 0; i < 9; i++ {
        for j := 0; j < 9; j++ {

            // initialize our candidates board
            candidates[i][j] = make(elementset)

            // if the tile is already occupied, nothing to do
            if board[i][j] != byte('.') {
                continue
            }

            subbox := coordsToSubbox(i, j)
            for el = 0; el < 9; el++ {
                if !rows[i][el] && !cols[j][el] && !subboxes[subbox][el] {
                    candidates[i][j].add(el)
                }
            }
        }
    }

    // primary loop for solving the puzzle
    for {

        // first priority: look for tiles with a single candidate
        res, res_board = lowHangingFruit(board, candidates)
        if res_board != nil && res > 0 {
            num_solved += res
            // 81 tiles solved means sudoku is solved. return
            if num_solved >= 81 {
                return
            }
            // if we got here, we're making progress but haven't solved it yet.
            // iterate the loop again
            continue
        } else if res_board == nil {
            break
        }

        // second priority: look for rows/columns/subboxes that have exactly 1
        // tile with a given candidate
        res, res_board = graspAtStraws(board, candidates)
        if res_board != nil && res > 0 {
            num_solved += res
            // 81 tiles solved means sudoku is solved. return
            if num_solved >= 81 {
                return
            }
            // if we got here, we're making progress but haven't solved it yet.
            // iterate the loop again
            continue
        } else if res_board == nil {
            break
        }

        // third priority: take random guesses
        res, res_board = shotInTheDark(board, candidates)
        if res_board != nil && res > 0 {
            copy(board, res_board)
            num_solved += res
            // 81 tiles solved means sudoku is solved. return
            if num_solved >= 81 {
                return
            }
            // if we got here, we're making progress but haven't solved it yet.
            // iterate the loop again
            continue
        } else if res_board == nil {
            break
        }

        break

    }

    if num_solved < 81 {
        // TODO remove the following
        fmt.Println("Board at end:")
        var converted_board [][]int8
        converted_board = make([][]int8, 9)
        for i := 0; i < 9; i++ {
            converted_board[i] = make([]int8, 9)
            for j := 0; j < 9; j++ {
                converted_board[i][j] = btoi[board[i][j]]+1
            }
        }
        for _, row := range converted_board {
            fmt.Println(row)
        }
    }

}
