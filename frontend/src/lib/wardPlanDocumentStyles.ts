/** Inline CSS for downloadable / printable ward goals documents. */
export const WARD_PLAN_STYLES = `
  @page {
    size: letter;
    margin: 0.7in 0.8in;
  }
  :root { color-scheme: light; }
  * { box-sizing: border-box; }
  html, body { height: 100%; }
  body {
    margin: 0;
    padding: 1.5rem 1rem 2rem;
    font-family: "Segoe UI", system-ui, -apple-system, sans-serif;
    font-size: 11pt;
    line-height: 1.5;
    color: #1b1812;
    background: #ece4cf;
  }
  .sheet-print-hint {
    max-width: 6.5in;
    margin: 0 auto 0.75rem;
    padding: 0.5rem 0.7rem;
    text-align: center;
    font-size: 9.5pt;
    line-height: 1.35;
    color: #1f3c63;
    background: #fffdf6;
    border: 1px solid #e2dac6;
    border-radius: 4px;
  }
  .sheet-print-hint strong { font-weight: 700; }
  .ward-goals-sheet {
    max-width: 6.5in;
    margin: 0 auto;
    padding: 0.45in 0.52in 0.48in;
    background: linear-gradient(165deg, #fdfcfa 0%, #f5f0e6 45%, #fffdf9 100%);
    border: 1px solid #cbc0a4;
    border-top: 5px solid #2b2b2b;
    border-radius: 3px;
    box-shadow: 0 10px 28px rgba(27, 24, 18, 0.12);
  }
  .ward-goals-sheet__header {
    margin: 0 0 0.28in;
    padding-bottom: 0.22in;
    border-bottom: 1px solid rgba(31, 60, 99, 0.12);
  }
  .ward-goals-sheet__eyebrow {
    margin: 0 0 0.1in;
    font-size: 7.5pt;
    font-weight: 800;
    letter-spacing: 0.22em;
    text-transform: uppercase;
    color: #8a6633;
  }
  .ward-goals-sheet__title {
    margin: 0 0 0.12in;
    font-family: Georgia, "Times New Roman", serif;
    font-size: 24pt;
    font-weight: 600;
    line-height: 1.12;
    letter-spacing: -0.015em;
    color: #0f0e0b;
  }
  .ward-goals-sheet__lead {
    margin: 0;
    font-size: 11pt;
    line-height: 1.45;
    color: #1b1812;
  }
  .ward-goals-sheet__lead strong { font-weight: 700; }
  .ward-goals-sheet__list {
    margin: 0;
    padding: 0;
    list-style: none;
  }
  .ward-goals-sheet__list li {
    position: relative;
    margin: 0 0 0.16in;
    padding: 0 0 0 0.24in;
    font-size: 11pt;
    line-height: 1.55;
    color: #1b1812;
    page-break-inside: avoid;
    break-inside: avoid;
  }
  .ward-goals-sheet__list li:last-child {
    margin-bottom: 0;
  }
  .ward-goals-sheet__list li::before {
    content: "";
    position: absolute;
    left: 0;
    top: 0.48em;
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: #b08745;
    box-shadow: 0 0 0 2px rgba(176, 135, 69, 0.22);
  }
  .ward-rich-line :is(p, ul, ol) {
    margin: 0.2rem 0;
  }
  .ward-rich-line :is(p, ul, ol):first-child {
    margin-top: 0;
  }
  .ward-rich-line :is(p, ul, ol):last-child {
    margin-bottom: 0;
  }
  .ward-rich-line strong,
  .ward-goals-sheet__list strong {
    font-weight: 700;
    color: #0f0e0b;
  }
  .ward-goals-sheet__quote {
    margin: 0.34in 0 0;
    padding: 0.14in 0 0.14in 0.18in;
    border-left: 3px solid #b08745;
    font-family: Georgia, "Times New Roman", serif;
    font-size: 11pt;
    font-style: italic;
    line-height: 1.45;
    color: #5b554a;
    page-break-inside: avoid;
    break-inside: avoid;
  }
  a { color: #1f3c63; }
  @media print {
    @page {
      size: letter;
      margin: 0.7in 0.8in;
    }
    html, body {
      height: auto;
    }
    body {
      padding: 0 !important;
      background: #fff !important;
    }
    .sheet-print-hint {
      display: none !important;
    }
    .ward-goals-sheet {
      max-width: none;
      margin: 0;
      padding: 0;
      border: none;
      border-top: 5px solid #2b2b2b;
      border-radius: 0;
      box-shadow: none;
      background: #fff;
      print-color-adjust: exact;
      -webkit-print-color-adjust: exact;
    }
    .ward-goals-sheet__header {
      border-bottom-color: rgba(31, 60, 99, 0.12);
    }
    .ward-goals-sheet__list li::before {
      print-color-adjust: exact;
      -webkit-print-color-adjust: exact;
    }
    .ward-goals-sheet__quote {
      border-left-color: #b08745;
    }
  }
`

export const WARD_PLAN_PRINT_HINT =
  'Before printing: paper size <strong>Letter</strong>, scale <strong>Default</strong>, and turn on <strong>Background graphics</strong> so the cream background and gold accents print correctly.'
