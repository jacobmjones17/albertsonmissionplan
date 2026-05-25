/** Inline CSS for the one-page monthly challenges sheet (download + print preview). */
export const CHALLENGES_SHEET_STYLES = `
  @page {
    size: 11in 8.5in;
    margin: 0.15in;
  }
  * { box-sizing: border-box; }
  html, body { height: 100%; }
  body {
    margin: 0;
    padding: 0.2in 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    font-family: "Segoe UI", system-ui, -apple-system, sans-serif;
    color: #1b1812;
    background: #d9d2c4;
  }
  .sheet-print-hint {
    width: 10.6in;
    max-width: calc(100% - 1rem);
    margin: 0 0 0.1in;
    padding: 0.08in 0.14in;
    text-align: center;
    font-size: 9.5pt;
    line-height: 1.35;
    color: #1f3c63;
    background: #fffdf6;
    border: 1px solid #e2dac6;
    border-radius: 4px;
    box-shadow: 0 1px 4px rgba(20, 18, 12, 0.08);
  }
  .sheet-print-hint strong { font-weight: 700; }
  .challenges-sheet {
    width: 10.6in;
    max-width: calc(100% - 1rem);
    height: 8in;
    padding: 0.08in 0.1in 0.1in;
    background: #fff;
    display: flex;
    flex-direction: column;
    box-shadow: 0 2px 18px rgba(20, 18, 12, 0.12);
  }
  .challenges-sheet__titlebar {
    flex-shrink: 0;
    text-align: center;
    margin: 0 0 0.06in;
    padding-bottom: 0.04in;
    border-bottom: 1px solid #e2dac6;
  }
  .challenges-sheet__title {
    margin: 0;
    font-family: Georgia, "Times New Roman", serif;
    font-size: 12pt;
    font-weight: 600;
    line-height: 1.1;
    color: #1f3c63;
  }
  .challenges-sheet__grid {
    flex: 1;
    min-height: 0;
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    grid-template-rows: repeat(3, minmax(0, 1fr));
    gap: 0.07in 0.08in;
  }
  .challenges-sheet__cell {
    min-width: 0;
    min-height: 0;
  }
  .challenges-sheet__card {
    height: 100%;
    min-height: 0;
    display: flex;
    flex-direction: column;
    background: #fffdf6;
    border: 1px solid #e2dac6;
    border-radius: 3px;
    overflow: hidden;
  }
  .challenges-sheet__art {
    position: relative;
    flex: 0 0 56%;
    min-height: 0;
    overflow: hidden;
    padding: 0.02in 0;
    background: linear-gradient(165deg, #f8f3ea 0%, #efe7d8 100%);
    border-bottom: 1px solid #ece4cf;
  }
  .challenges-sheet__img {
    display: block;
    width: 100%;
    height: 100%;
    object-fit: contain;
    object-position: center;
  }
  .challenges-sheet__month {
    position: absolute;
    top: 0.04in;
    left: 0.05in;
    z-index: 1;
    padding: 0.015in 0.07in;
    font-size: 7pt;
    font-weight: 700;
    letter-spacing: 0.08em;
    text-transform: uppercase;
    color: #1f3c63;
    background: rgba(255, 253, 246, 0.95);
    border: 1px solid rgba(31, 60, 99, 0.22);
    border-radius: 999px;
  }
  .challenges-sheet__body {
    flex: 1;
    min-height: 0;
    padding: 0.04in 0.06in 0.04in;
    overflow: hidden;
  }
  .challenges-sheet__goal {
    margin: 0 0 0.025in;
    font-size: 7pt;
    line-height: 1.18;
    color: #1b1812;
  }
  .challenges-sheet__goal strong { font-weight: 700; }
  .challenges-sheet__list {
    margin: 0;
    padding: 0;
    list-style: none;
    border-top: 1px solid #ece4cf;
  }
  .challenges-sheet__list li {
    position: relative;
    margin: 0;
    padding: 0.014in 0 0.014in 0.09in;
    font-size: 6.75pt;
    line-height: 1.16;
    color: #3a3630;
  }
  .challenges-sheet__list li::before {
    content: "";
    position: absolute;
    left: 0;
    top: 0.057in;
    width: 3px;
    height: 3px;
    background: #b08745;
    border-radius: 1px;
  }
  .challenges-sheet__list li + li {
    border-top: 1px solid #f0ebe0;
  }
  .challenges-sheet__list strong { color: #1f3c63; font-weight: 700; }
  a { color: #1f3c63; text-decoration: none; }
  @media print {
    @page {
      size: 11in 8.5in;
      margin: 0.15in;
    }
    html {
      height: 100%;
    }
    body {
      width: auto;
      height: 100%;
      margin: 0 !important;
      padding: 0 !important;
      background: #fff !important;
      display: flex !important;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      overflow: hidden;
    }
    .sheet-print-hint {
      display: none !important;
    }
    .challenges-sheet {
      width: 10.6in !important;
      max-width: none !important;
      height: 8in !important;
      margin: 0 !important;
      box-shadow: none !important;
      flex-shrink: 0;
      page-break-after: avoid;
      page-break-inside: avoid;
      break-inside: avoid;
    }
    .challenges-sheet__card {
      print-color-adjust: exact;
      -webkit-print-color-adjust: exact;
    }
  }
`

export const CHALLENGES_PRINT_HINT =
  'Before printing: set <strong>Layout to Landscape</strong>, paper size to <strong>Letter</strong>, and scale to <strong>100%</strong>. If the preview looks tiny with blank space below, the layout is still set to Portrait — switch it to Landscape.'
