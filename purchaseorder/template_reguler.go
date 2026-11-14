package purchaseorder

import "io"

// renderReguler renders a Surat Pesanan Reguler: the order letter for obat
// bebas, bebas terbatas, keras and alat kesehatan.
func renderReguler(w io.Writer, doc Document) error {
	l := newLetter(doc)

	l.writeLetterhead()
	l.writeTitle("SURAT PESANAN")
	l.writeRecipient()

	l.pdf.Ln(3)
	l.writeParagraph("Dengan hormat,\nbersama ini kami mengajukan pesanan obat sebagai berikut:")

	const closingHeight = 4 + bodyLineHeight
	l.writeItems(closingHeight + signatureHeight(doc.Signer != nil))

	l.pdf.Ln(4)
	l.writeParagraph("Atas perhatian dan kerja samanya, kami ucapkan terima kasih.")
	l.writeSignature()

	return l.output(w)
}
