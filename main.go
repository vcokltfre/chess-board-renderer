package main

import (
	"embed"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo"
)

//go:embed static
var static embed.FS

func loadImage(filename string) image.Image {
	f, err := static.Open("static/" + filename)
	if err != nil {
		panic(err)
	}

	img, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}

	return img
}

var (
	WhitePawnImage   = loadImage("pawn_white.png")
	WhiteKnightImage = loadImage("knight_white.png")
	WhiteBishopImage = loadImage("bishop_white.png")
	WhiteRookImage   = loadImage("rook_white.png")
	WhiteQueenImage  = loadImage("queen_white.png")
	WhiteKingImage   = loadImage("king_white.png")
	BlackPawnImage   = loadImage("pawn_black.png")
	BlackKnightImage = loadImage("knight_black.png")
	BlackBishopImage = loadImage("bishop_black.png")
	BlackRookImage   = loadImage("rook_black.png")
	BlackQueenImage  = loadImage("queen_black.png")
	BlackKingImage   = loadImage("king_black.png")
)

var FEN = regexp.MustCompile(`^([rnbqkpRNBQKP1-8]{1,8}/){7}[rnbqkpRNBQKP1-8]{1,8}$`)

type Piece int

const (
	Empty Piece = iota
	WhitePawn
	WhiteKnight
	WhiteBishop
	WhiteRook
	WhiteQueen
	WhiteKing
	BlackPawn
	BlackKnight
	BlackBishop
	BlackRook
	BlackQueen
	BlackKing
)

var pieceImages = map[Piece]image.Image{
	WhitePawn:   WhitePawnImage,
	WhiteKnight: WhiteKnightImage,
	WhiteBishop: WhiteBishopImage,
	WhiteRook:   WhiteRookImage,
	WhiteQueen:  WhiteQueenImage,
	WhiteKing:   WhiteKingImage,
	BlackPawn:   BlackPawnImage,
	BlackKnight: BlackKnightImage,
	BlackBishop: BlackBishopImage,
	BlackRook:   BlackRookImage,
	BlackQueen:  BlackQueenImage,
	BlackKing:   BlackKingImage,
}

var pieceChars = map[rune]Piece{
	'P': WhitePawn,
	'N': WhiteKnight,
	'B': WhiteBishop,
	'R': WhiteRook,
	'Q': WhiteQueen,
	'K': WhiteKing,
	'p': BlackPawn,
	'n': BlackKnight,
	'b': BlackBishop,
	'r': BlackRook,
	'q': BlackQueen,
	'k': BlackKing,
}

type Board struct {
	Pieces [8][8]Piece
}

func validate(board string) (*Board, error) {
	if !FEN.MatchString(board) {
		return nil, echo.NewHTTPError(400, "Invalid FEN")
	}

	result := &Board{
		Pieces: [8][8]Piece{},
	}

	segments := strings.Split(board, "/")
	for row, segment := range segments {
		rowPieces := []Piece{}
		for _, char := range segment {
			piece := pieceChars[char]
			if piece == Empty {
				blanks, _ := strconv.Atoi(string(char))
				for i := 0; i < blanks; i++ {
					rowPieces = append(rowPieces, Empty)
				}
				continue
			}

			rowPieces = append(rowPieces, piece)
		}

		if len(rowPieces) != 8 {
			return nil, echo.NewHTTPError(400, "Invalid FEN")
		}

		copy(result.Pieces[row][:], rowPieces)
	}

	return result, nil
}

func render(board string, c echo.Context) error {
	start := time.Now()

	b, err := validate(board)
	if err != nil {
		return c.String(400, err.Error())
	}

	img := image.NewRGBA(image.Rect(0, 0, 512, 512))
	draw.Draw(img, img.Bounds(), image.White, image.Point{}, draw.Src)

	for row := 0; row < 8; row++ {
		for column := 0; column < 8; column++ {
			piece := b.Pieces[row][column]
			tileColour := (row + column) % 2

			if tileColour == 0 {
				draw.Draw(img, image.Rect(column*64, row*64, (column*64)+64, (row*64)+64), image.NewUniform(color.RGBA{
					R: 0x4f,
					G: 0x4f,
					B: 0x4f,
					A: 0xff,
				}), image.Point{}, draw.Src)
			}

			if piece == Empty {
				continue
			}

			pieceImage := pieceImages[piece]

			draw.Draw(img, image.Rect(column*64, row*64, (column*64)+64, (row*64)+64), pieceImage, image.Point{}, draw.Over)
		}
	}

	processingTime := time.Since(start)

	c.Response().Header().Set("Content-Type", "image/png")
	c.Response().Header().Set("X-Processing-Time", processingTime.String())
	png.Encode(c.Response().Writer, img)

	fmt.Printf("Rendered board in %s\n", processingTime)

	return nil
}

func main() {
	e := echo.New()

	e.GET("/render", func(c echo.Context) error {
		board := c.QueryParam("board")

		return render(board, c)
	})

	bind := ":8080"
	if val, ok := os.LookupEnv("BIND"); ok {
		bind = val
	}

	if err := e.Start(bind); err != nil {
		e.Logger.Fatal(err)
	}
}
