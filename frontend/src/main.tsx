import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { BrowserRouter } from 'react-router-dom'
import { BootstrapProvider } from './BootstrapContext'
import { App } from './App'
import './index.css'
import './site.css'
import './print.css'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <BrowserRouter>
      <BootstrapProvider>
        <App />
      </BootstrapProvider>
    </BrowserRouter>
  </StrictMode>,
)
