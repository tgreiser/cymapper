package app

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
    "errors"

	"github.com/g3n/engine/gui"
	"github.com/g3n/engine/gui/assets/icon"
	"github.com/g3n/engine/math32"
)

type FileSelect struct {
	gui.Panel
	title    *gui.Label
	path     *gui.Label
	filename *gui.Edit
	list     *gui.List
	bok      *gui.Button
	bcan     *gui.Button
}

func NewFileSelect(width, height float32) (*FileSelect, error) {

	fs := new(FileSelect)
	fs.Panel.Initialize(width, height)
	fs.SetBorders(2, 2, 2, 2)
	fs.SetPaddings(4, 4, 4, 4)
	fs.SetColor(math32.NewColor("White"))
	fs.SetVisible(false)
	fs.SetBounded(false)

	// Set vertical box layout for the whole panel
	l := gui.NewVBoxLayout()
	l.SetSpacing(4)
	fs.SetLayout(l)

	fs.title = gui.NewLabel("title")
	fs.title.SetColor(math32.NewColor("Black"))
	fs.Add(fs.title)

	// Creates path label
	fs.path = gui.NewLabel("path")
	fs.path.SetText("test")
	fs.path.SetColor(math32.NewColor("Black"))
	fs.Add(fs.path)

	// Creates list
	fs.list = gui.NewVList(0, 0)
	fs.list.SetLayoutParams(&gui.VBoxLayoutParams{Expand: 5, AlignH: gui.AlignWidth})
	fs.list.Subscribe(gui.OnChange, func(evname string, ev interface{}) {
		fs.onSelect()
	})
	fs.Add(fs.list)

	fs.filename = gui.NewEdit(400, "")
	fs.path.SetColor(math32.NewColor("White"))
	fs.Add(fs.filename)

	// Button container panel
	bc := gui.NewPanel(0, 0)
	bcl := gui.NewHBoxLayout()
	bcl.SetAlignH(gui.AlignWidth)
	bc.SetLayout(bcl)
	bc.SetLayoutParams(&gui.VBoxLayoutParams{Expand: 1, AlignH: gui.AlignWidth})
	fs.Add(bc)

	// Creates OK button
	fs.bok = gui.NewButton("OK")
	fs.bok.SetLayoutParams(&gui.HBoxLayoutParams{Expand: 0, AlignV: gui.AlignCenter})
	fs.bok.Subscribe(gui.OnClick, func(evname string, ev interface{}) {
		fs.Dispatch("OnOK", nil)
	})
	bc.Add(fs.bok)

	// Creates Cancel button
	fs.bcan = gui.NewButton("Cancel")
	fs.bcan.SetLayoutParams(&gui.HBoxLayoutParams{Expand: 0, AlignV: gui.AlignCenter})
	fs.bcan.Subscribe(gui.OnClick, func(evname string, ev interface{}) {
		fs.Dispatch("OnCancel", nil)
	})
	bc.Add(fs.bcan)

	// Sets initial directory
	path, err := os.Getwd()
	if err != nil {
		return nil, err
	} else {
		fs.SetPath(path)
	}
	return fs, nil
}

// Show shows or hide the file selection dialog
func (fs *FileSelect) Show(show bool) {

	if show {
		fs.SetVisible(true)
        fs.SetPath(fs.path.Text())
		parent := fs.Parent().(gui.IPanel).GetPanel()
		px := (parent.Width() - fs.Width()) / 2
		py := (parent.Height() - fs.Height()) / 2
		fs.SetPosition(px, py)
	} else {
		fs.SetVisible(false)
	}
}

func (fs *FileSelect) SetTitle(title string) {
	fs.title.SetText(title)
}

func (fs *FileSelect) SetFilename(name string) {
	fs.filename.SetText(name)
}

func (fs *FileSelect) SetPath(path string) error {

	// Open path file or dir
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Checks if it is a directory
	files, err := f.Readdir(0)
	if err != nil {
		return err
	}
	fs.path.SetText(path)

	// Sort files by name
	sort.Sort(listFileInfo(files))

	// Reads directory contents and loads into the list
	fs.list.Clear()
	// Adds previous directory
	prev := gui.NewImageLabel("..")
	prev.SetIcon(icon.FolderOpen)
	fs.list.Add(prev)
	// Adds directory files
	for i := 0; i < len(files); i++ {
		item := gui.NewImageLabel(files[i].Name())
		n := strings.ToLower(files[i].Name())
		if files[i].IsDir() {
			item.SetIcon(icon.FolderOpen)
			fs.list.Add(item)
		} else if n[len(n)-4:] == ".tsv" {
			item.SetIcon(icon.InsertPhoto)
			fs.list.Add(item)
		}

	}
	return nil
}

func (fs *FileSelect) Selected() (string, error) {
    fileNameLowerBox := fs.filename.Text()
    if len(fileNameLowerBox) == 0 {
        return "", errors.New("file not selected")
    }
    return filepath.Join(fs.path.Text(), fileNameLowerBox), nil
}

func (fs *FileSelect) onSelect() {

	// Get selected image label and its txt
	sel := fs.list.Selected()[0]
	label := sel.(*gui.ImageLabel)
	text := label.Text()

	// Checks if previous directory
	if text == ".." {
		dir, _ := filepath.Split(fs.path.Text())
		fs.SetPath(filepath.Dir(dir))
		return
	}

	// Checks if it is a directory
	path := filepath.Join(fs.path.Text(), text)
	s, err := os.Stat(path)
	if err != nil {
		panic(err)
		return
	}
	if s.IsDir() {
		fs.SetPath(path)
	} else {
		fs.SetFilename(text)
	}
}

// For sorting array of FileInfo by Name
type listFileInfo []os.FileInfo

func (fi listFileInfo) Len() int      { return len(fi) }
func (fi listFileInfo) Swap(i, j int) { fi[i], fi[j] = fi[j], fi[i] }
func (fi listFileInfo) Less(i, j int) bool {

	return fi[i].Name() < fi[j].Name()
}
