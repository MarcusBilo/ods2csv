// Version-: 05-11-2017

//////////////////////////////////////contents.xml format///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//<office:spreadsheet>                                                                          							                                      //
//                                                                                                           								                      //
// <table:table table:name="name" table:style-name="ta1">                                                                                                     		                      //
//                                                                                                          								                      //
//      <table:table-row table:number-rows-repeated="2" table:style-name="ro1">                              								                      //
//                                                                                                           								                      //
//      <table:table-cell table:formula="of:=3*[.B2]"  table:number-columns-repeated="2" table:style-name="ce1" office:value-type="string" calcext:value-type="string" office:date-value="" > //
//      <text:p>SrNo<text:span text:style-name="T1">gj</text:span><text:s text:c="10"/><text:s text:c="10"/>gh</text:p>                                                                       //
//      </table:table-cell>                                                                                  								                      //
//                                                                                                                                                  			                      //
// </table:table-row>                                                                                        								                      //
//                                                                                                           								                      //
// </office:spreadsheet>                                                                                     								                      //
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

package ods

import (
	"archive/zip"
	//	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/LIJUCHACKO/XmlDB"
)

type Cell struct {
	Type      string //Type float,string ...    ( office:value-type= )
	Value     string //Value                    ( office:value= )
	DateValue string //DateValue                ( office:date-value= )
	Formula   string //formula                  ( table:formula= )
	Text      string //Text

}

type Row struct {
	Cells []Cell
}

type Sheet struct {
	Name string
	Rows []Row
}

type Odsfile struct {
	Sheets []Sheet
}
type Result struct {
	Sheeti Sheet
	err    error
}

func ReplaceHTMLSpecialEntities(input string) string {
var b strings.Builder
	b.Grow(len(input))
	i := 0
	for i < len(input) {
		if input[i] == '&' {
			switch {
			case strings.HasPrefix(input[i:], "&amp;"):
				b.WriteString("&")
				i += len("&amp;")
			case strings.HasPrefix(input[i:], "&lt;"):
				b.WriteString("<")
				i += len("&lt;")
			case strings.HasPrefix(input[i:], "&gt;"):
				b.WriteString(">")
				i += len("&gt;")
			case strings.HasPrefix(input[i:], "&quot;"):
				b.WriteString("\"")
				i += len("&quot;")
			case strings.HasPrefix(input[i:], "&lsquo;"):
				b.WriteString("‘")
				i += len("&lsquo;")
			case strings.HasPrefix(input[i:], "&rsquo;"):
				b.WriteString("’")
				i += len("&rsquo;")
			case strings.HasPrefix(input[i:], "&tilde;"):
				b.WriteString("~")
				i += len("&tilde;")
			case strings.HasPrefix(input[i:], "&ndash;"):
				b.WriteString("–")
				i += len("&ndash;")
			case strings.HasPrefix(input[i:], "&mdash;"):
				b.WriteString("—")
				i += len("&mdash;")
			case strings.HasPrefix(input[i:], "&apos;"):
				b.WriteString("'")
				i += len("&apos;")
			default:
				b.WriteByte(input[i])
				i++
			}
		} else {
			b.WriteByte(input[i])
			i++
		}
	}
	return b.String()
}
func ReadSheet(DB *xmlDB.Database, spreadsheet int) (Sheet, error) {
	tablename := xmlDB.GetNodeAttribute(DB, spreadsheet, "table:name")
	//fmt.Printf("\nStarted %s\n", tablename)
	csvrows := []Row{}
	rows, _ := xmlDB.GetNode(DB, spreadsheet, "table:table-row")
	blankrows := 0

	for _, row := range rows {
		rowisblank := true
		row_repetitionTXT := xmlDB.GetNodeAttribute(DB, row, "table:number-rows-repeated")
		row_repetition := 1
		if len(strings.TrimSpace(row_repetitionTXT)) > 0 {
			value, err := strconv.Atoi(row_repetitionTXT)
			if err != nil {
				return Sheet{tablename, csvrows}, err
			}
			row_repetition = value
		}
		csvcells := []Cell{}
		Cells, _ := xmlDB.GetNode(DB, row, "table:table-cell")
		blankcells := 0
		for _, cell := range Cells {
			cell_repetitionTXT := xmlDB.GetNodeAttribute(DB, cell, "table:number-columns-repeated")
			cell_repetition := 1
			if len(strings.TrimSpace(cell_repetitionTXT)) > 0 {
				value, err := strconv.Atoi(cell_repetitionTXT)
				if err != nil {
					return Sheet{tablename, csvrows}, err
				}
				cell_repetition = value
			}
			celltype := xmlDB.GetNodeAttribute(DB, cell, "office:value-type")
			celldatevalue := xmlDB.GetNodeAttribute(DB, cell, "office:date-value")
			cellvalue := xmlDB.GetNodeAttribute(DB, cell, "office:value")
			cellformula := xmlDB.GetNodeAttribute(DB, cell, "table:formula")

			Cell_paras, _ := xmlDB.GetNode(DB, cell, "text:p")
			celltext := ""
			for _, Cell_para := range Cell_paras {
				childnodes := xmlDB.ChildNodes(DB, Cell_para)
				if len(childnodes) > 0 {
					for _, child := range childnodes {
						nodeName := xmlDB.GetNodeName(DB, child)
						if nodeName == "text:s" {
							cell_space := 1
							cell_spaceTXT := xmlDB.GetNodeAttribute(DB, child, "text:c")
							if len(strings.TrimSpace(cell_spaceTXT)) > 0 {
								value, err := strconv.Atoi(cell_spaceTXT)
								if err != nil {
									return Sheet{tablename, csvrows}, err
								}
								cell_space = value
							}
							for {
								if cell_space == 0 {
									break
								}
								celltext = celltext + " "
								cell_space--
							}
						}
						celltext = celltext + xmlDB.GetNodeValue(DB, child)
					}
				} else {
					if len(celltext) > 0 {
						celltext = celltext + "\n" + xmlDB.GetNodeValue(DB, Cell_para)
					} else {
						celltext = xmlDB.GetNodeValue(DB, Cell_para)
					}

				}
			}

			celltext = ReplaceHTMLSpecialEntities(celltext)
			cellvalue = ReplaceHTMLSpecialEntities(cellvalue)
			if len(celltext) == 0 {
				blankcells = blankcells + cell_repetition
			} else {
				//insert blankcells before newcells
				for {
					if blankcells == 0 {
						break
					}
					csvcells = append(csvcells, Cell{"", "", "", "", ""})
					blankcells--

				}
				for {
					if cell_repetition == 0 {
						break
					}
					csvcells = append(csvcells, Cell{celltype, cellvalue, celldatevalue, cellformula, celltext})
					cell_repetition--
					rowisblank = false
				}
			}

		}
		if rowisblank {
			blankrows = blankrows + row_repetition
		} else {
			//insert blankrows before newrows
			for {
				if blankrows == 0 {
					break
				}
				csvrows = append(csvrows, Row{[]Cell{}})
				blankrows--
			}
			for {
				if row_repetition == 0 {
					break
				}
				csvrows = append(csvrows, Row{csvcells})
				row_repetition--
			}

		}

	}
	return Sheet{tablename, csvrows}, nil
}
func ReadSheetThread(DB *xmlDB.Database, spreadsheet int, chres chan Result) {
	res := new(Result)
	Sheeti, err := ReadSheet(DB, spreadsheet)
	res.Sheeti = Sheeti
	res.err = err
	chres <- *res
}
func ReadODSFile(odsfilename string) (Odsfile, error) {
	var odsfileContents Odsfile
	var DB *xmlDB.Database = new(xmlDB.Database)
	DB.MaxNooflines = 999999
	DB.Debug_enabled = false
	r, err := zip.OpenReader(odsfilename)
	if err != nil {
		return odsfileContents, err
	}
	defer r.Close()

	for _, f := range r.File {
		if f.Name == "content.xml" {
			rc, fileerr1 := f.Open()
			if fileerr1 != nil {
				return odsfileContents, fileerr1
			}
			xmlfile, fileerr := ioutil.ReadAll(rc)
			if fileerr != nil {
				return odsfileContents, fileerr
			}

			xmlline := string(xmlfile)
			xmllines := strings.Split(xmlline, "\n")
			xmlDB.Load_dbcontent(DB, xmllines)
			//fmt.Printf("\ncontent.xml loaded to xmldb\n")
			csvSpreadSheets := []Sheet{}
			spreadSheets, _ := xmlDB.GetNode(DB, 0, "office:body/office:spreadsheet/table:table")
			chan1 := make(chan Result, 8)
			for _, spreadsheet := range spreadSheets {
				go ReadSheetThread(DB, spreadsheet, chan1)
			}
			read := 0
			for {
				if read >= len(spreadSheets) {
					break
				}
				result := <-chan1
				csvSpreadSheets = append(csvSpreadSheets, result.Sheeti)
				if result.err != nil {
					err = result.err
					return odsfileContents, err
				}
				read++
			}
			close(chan1)
			odsfileContents.Sheets = csvSpreadSheets
			return odsfileContents, err
		}
	}

	return odsfileContents, err
}
