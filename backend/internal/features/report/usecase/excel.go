package usecase

import (
	"bytes"
	"fmt"
	"strings"

	reportdomain "apihorpug/internal/features/report/domain"

	"github.com/xuri/excelize/v2"
)

// The workbook is for the dormitory owner's bookkeeping, so it is labelled in
// Thai like the rest of the tenant-facing documents.
var (
	methodLabels = map[string]string{
		"cash": "เงินสด", "transfer": "โอนเงิน", "credit_card": "บัตรเครดิต", "other": "อื่น ๆ", "deposit": "หักจากเงินประกัน",
	}
	itemTypeLabels = map[string]string{
		"rent": "ค่าเช่าห้อง", "electricity": "ค่าไฟฟ้า", "water": "ค่าน้ำประปา", "other": "อื่น ๆ",
	}
	categoryLabels = map[string]string{
		"maintenance": "ซ่อมบำรุง", "utility": "สาธารณูปโภค", "salary": "เงินเดือน", "supplies": "วัสดุสิ้นเปลือง", "other": "อื่น ๆ",
	}
	statusLabels = map[string]string{
		"unpaid": "ยังไม่ชำระ", "paid": "ชำระแล้ว", "overdue": "เกินกำหนด", "cancelled": "ยกเลิก",
	}
	thaiMonths = [...]string{"", "มกราคม", "กุมภาพันธ์", "มีนาคม", "เมษายน", "พฤษภาคม", "มิถุนายน",
		"กรกฎาคม", "สิงหาคม", "กันยายน", "ตุลาคม", "พฤศจิกายน", "ธันวาคม"}
)

func label(labels map[string]string, key string) string {
	if l, ok := labels[key]; ok {
		return l
	}
	return key
}

// sheetWriter appends rows to one sheet, tracking the next row.
type sheetWriter struct {
	f     *excelize.File
	sheet string
	row   int
	money int
	bold  int
	err   error
}

func (w *sheetWriter) put(values []any, style int) {
	if w.err != nil {
		return
	}
	w.row++
	cell, _ := excelize.CoordinatesToCellName(1, w.row)
	if w.err = w.f.SetSheetRow(w.sheet, cell, &values); w.err != nil {
		return
	}
	if style != 0 {
		end, _ := excelize.CoordinatesToCellName(len(values), w.row)
		w.err = w.f.SetCellStyle(w.sheet, cell, end, style)
	}
}

// moneyColumns applies the number format to the given 1-based columns.
func (w *sheetWriter) moneyColumns(cols ...int) {
	for _, col := range cols {
		if w.err != nil {
			return
		}
		name, _ := excelize.ColumnNumberToName(col)
		w.err = w.f.SetColStyle(w.sheet, name, w.money)
	}
}

func (w *sheetWriter) widths(widths ...float64) {
	for i, width := range widths {
		if w.err != nil {
			return
		}
		name, _ := excelize.ColumnNumberToName(i + 1)
		w.err = w.f.SetColWidth(w.sheet, name, name, width)
	}
}

func buildWorkbook(rep reportdomain.MonthlyReport, d reportdomain.Details) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	money, err := f.NewStyle(&excelize.Style{NumFmt: 4}) // #,##0.00
	if err != nil {
		return nil, err
	}
	bold, err := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	if err != nil {
		return nil, err
	}
	title, err := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Size: 14}})
	if err != nil {
		return nil, err
	}

	period := fmt.Sprintf("%s %d", thaiMonths[rep.Month], rep.Year+543)

	// Summary
	const summary = "สรุป"
	if err := f.SetSheetName("Sheet1", summary); err != nil {
		return nil, err
	}
	s := &sheetWriter{f: f, sheet: summary, money: money, bold: bold}
	s.widths(36, 18)
	s.moneyColumns(2)
	s.put([]any{"รายงานประจำเดือน " + period}, title)
	s.put(nil, 0)
	s.put([]any{"รายรับ (รับชำระในเดือน)", rep.Income.Total}, bold)
	for _, a := range rep.Income.ByMethod {
		s.put([]any{"   " + label(methodLabels, a.Key), a.Amount}, 0)
	}
	s.put([]any{"ค่าใช้จ่าย", rep.Expenses.Total}, bold)
	for _, a := range rep.Expenses.ByCategory {
		s.put([]any{"   " + label(categoryLabels, a.Key), a.Amount}, 0)
	}
	s.put([]any{"รายรับสุทธิ (รายรับ − ค่าใช้จ่าย)", rep.Net}, bold)
	s.put(nil, 0)
	s.put([]any{fmt.Sprintf("ใบแจ้งหนี้งวด %s (%d ใบ)", period, rep.Billing.InvoiceCount), rep.Billing.Billed}, bold)
	for _, a := range rep.Billing.ByItemType {
		s.put([]any{"   " + label(itemTypeLabels, a.Key), a.Amount}, 0)
	}
	s.put([]any{"   เก็บได้แล้ว", rep.Billing.Collected}, 0)
	s.put([]any{"   ค้างชำระ", rep.Billing.Outstanding}, 0)
	s.put([]any{fmt.Sprintf("ยอดค้างชำระสะสมทุกงวด ณ วันออกรายงาน (%d ใบ)", rep.Arrears.Count), rep.Arrears.Amount}, bold)
	s.put([]any{fmt.Sprintf("ห้องที่มีผู้เช่า (ปัจจุบัน) %d / %d ห้อง", rep.Occupancy.Occupied, rep.Occupancy.Rooms)}, 0)
	if len(rep.Dormitories) > 1 {
		s.put(nil, 0)
		s.put([]any{"แยกตามหอพัก", "รายรับ", "ค่าใช้จ่าย", "สุทธิ", "ยอดบิลงวดนี้", "ค้างชำระงวดนี้"}, bold)
		s.moneyColumns(2, 3, 4, 5, 6)
		for _, row := range rep.Dormitories {
			s.put([]any{row.Name, row.Income, row.Expenses, row.Net, row.Billed, row.Outstanding}, 0)
		}
	}
	if s.err != nil {
		return nil, s.err
	}

	// Payments
	p := &sheetWriter{f: f, sheet: "รับชำระ", money: money, bold: bold}
	if _, err := f.NewSheet(p.sheet); err != nil {
		return nil, err
	}
	p.widths(14, 12, 20, 10, 24, 10, 22, 14, 10)
	p.moneyColumns(8)
	p.put([]any{"เลขที่ใบเสร็จ", "วันที่", "หอพัก", "ห้อง", "ผู้เช่า", "งวด", "ช่องทาง", "จำนวนเงิน", "สถานะ"}, bold)
	for _, l := range d.Payments {
		methods := ""
		for i, m := range strings.Split(l.Methods, ",") {
			if i > 0 {
				methods += ", "
			}
			methods += label(methodLabels, strings.TrimSpace(m))
		}
		status := ""
		if l.Voided {
			status = "ยกเลิก"
		}
		p.put([]any{l.ReceiptNo, l.PaymentDate.Format("02/01/2006"), l.DormitoryName, l.RoomNumber, l.TenantName,
			fmt.Sprintf("%02d/%d", l.PeriodMonth, l.PeriodYear), methods, l.Amount, status}, 0)
	}
	if p.err != nil {
		return nil, p.err
	}

	// Invoices
	inv := &sheetWriter{f: f, sheet: "ใบแจ้งหนี้", money: money, bold: bold}
	if _, err := f.NewSheet(inv.sheet); err != nil {
		return nil, err
	}
	inv.widths(20, 10, 24, 12, 14, 14, 14, 12)
	inv.moneyColumns(5, 6, 7)
	inv.put([]any{"หอพัก", "ห้อง", "ผู้เช่า", "ครบกำหนด", "ยอดบิล", "ชำระแล้ว", "คงค้าง", "สถานะ"}, bold)
	for _, l := range d.Invoices {
		owed := l.Total - l.Paid
		if owed < 0 || l.Status == "cancelled" {
			owed = 0
		}
		inv.put([]any{l.DormitoryName, l.RoomNumber, l.TenantName, l.DueDate.Format("02/01/2006"),
			l.Total, l.Paid, round2(owed), label(statusLabels, l.Status)}, 0)
	}
	if inv.err != nil {
		return nil, inv.err
	}

	// Expenses
	e := &sheetWriter{f: f, sheet: "ค่าใช้จ่าย", money: money, bold: bold}
	if _, err := f.NewSheet(e.sheet); err != nil {
		return nil, err
	}
	e.widths(12, 20, 18, 36, 14)
	e.moneyColumns(5)
	e.put([]any{"วันที่", "หอพัก", "หมวด", "รายละเอียด", "จำนวนเงิน"}, bold)
	for _, l := range d.Expenses {
		e.put([]any{l.ExpenseDate.Format("02/01/2006"), l.DormitoryName, label(categoryLabels, l.Category), l.Description, l.Amount}, 0)
	}
	if e.err != nil {
		return nil, e.err
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
