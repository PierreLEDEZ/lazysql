package views

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lazysql/lazysql/pkg/db"
	"github.com/lazysql/lazysql/pkg/tui/components"
	"github.com/lazysql/lazysql/pkg/tui/styles"
)

type ResultsModel struct {
	table    components.TableModel
	err      string
	elapsed  time.Duration
	rowCount int
	execMsg  string
	active   bool
	Width    int
	Height   int
}

func NewResults() ResultsModel {
	return ResultsModel{
		table: components.NewTable(),
	}
}

func (r ResultsModel) Init() tea.Cmd {
	return nil
}

func (r ResultsModel) Update(msg tea.Msg) (ResultsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case QueryResultMsg:
		if msg.Err != nil {
			r.err = msg.Err.Error()
			r.execMsg = ""
			return r, nil
		}
		r.err = ""
		r.execMsg = ""
		r.elapsed = msg.Result.Elapsed
		r.rowCount = msg.Result.RowCount
		r.setResultData(msg.Result)
		return r, nil
	case ExecResultMsg:
		if msg.Err != nil {
			r.err = msg.Err.Error()
			r.execMsg = ""
			return r, nil
		}
		r.err = ""
		r.execMsg = fmt.Sprintf("Query executed: %d rows affected (%s)", msg.Result.RowsAffected, msg.Result.Elapsed.Round(time.Millisecond))
		return r, nil
	}

	var cmd tea.Cmd
	r.table, cmd = r.table.Update(msg)
	return r, cmd
}

func (r ResultsModel) View() string {
	var titleText string
	if r.rowCount > 0 {
		titleText = fmt.Sprintf("Results (%d rows, %s)", r.rowCount, r.elapsed.Round(time.Millisecond))
	} else {
		titleText = "Results"
	}
	title := styles.PanelTitle(5, titleText, r.active)

	if r.err != "" {
		return fmt.Sprintf("%s\n%s", title, styles.ErrorText.Render("  Error: "+r.err))
	}
	if r.execMsg != "" {
		return fmt.Sprintf("%s\n%s", title, styles.SuccessText.Render("  "+r.execMsg))
	}

	if r.rowCount == 0 && r.execMsg == "" && r.err == "" {
		empty := styles.MutedText.Render(
			"  Select a table or write a query.\n" +
				"  ctrl+e to execute. ? for help.")
		return fmt.Sprintf("%s\n%s", title, empty)
	}

	return fmt.Sprintf("%s\n%s", title, r.table.View())
}

func (r *ResultsModel) SetSize(w, h int) {
	r.Width = w
	r.Height = h
	r.table.SetSize(w, h-1)
}

func (r *ResultsModel) Focus() {
	r.table.Focus()
}

func (r *ResultsModel) Blur() {
	r.table.Blur()
}

func (r *ResultsModel) SetActive(v bool) { r.active = v }

func (r *ResultsModel) setResultData(result *db.Result) {
	rows := make([][]string, len(result.Rows))
	for i, row := range result.Rows {
		strRow := make([]string, len(row))
		for j, cell := range row {
			strRow[j] = db.StringVal(cell)
		}
		rows[i] = strRow
	}
	r.table.SetData(result.Columns, rows)
}
