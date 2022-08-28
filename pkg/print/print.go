package print

import (
	"bytes"

	"github.com/mect/go-escpos"
	"github.com/meyskens/m-planner/pkg/db"
)

type bufferToRWC struct {
	u *bytes.Buffer
}

func (u *bufferToRWC) Read(p []byte) (int, error) {
	return u.u.Read(p)
}

func (u *bufferToRWC) Write(p []byte) (int, error) {
	return u.u.Write(p)
}

func (u *bufferToRWC) Close() error { // fake this out
	return nil
}

func PrintIdeaList(user string, ideas []db.Idea) ([]db.PrintJob, error) {
	data := bytes.NewBuffer(nil)

	p, err := escpos.NewPrinterByRW(&bufferToRWC{
		u: data,
	})

	if err != nil {
		return nil, err
	}

	p.Init()       // start
	p.Smooth(true) // use smootth printing
	p.Size(2, 2)   // set font size
	p.PrintLn("Ideas list")
	p.Size(1, 1)
	p.PrintLn("-----------")
	p.PrintLn("")

	p.Size(2, 1)
	for _, idea := range ideas {
		p.Print("* ")
		p.PrintLn(idea.Description)
	}

	p.PrintLn("")
	p.PrintLn("")
	p.Size(1, 1)
	p.PrintLn("Powered by M-Planner")

	return []db.PrintJob{
		{
			User:       user,
			EscposData: data.Bytes(),
		},
	}, nil
}
