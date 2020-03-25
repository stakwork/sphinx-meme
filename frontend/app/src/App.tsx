import React from 'react'
import {useObserver} from 'mobx-react-lite'
import {useStores} from './store'
import './App.css'
import dateformat from 'dateformat'
import Header from './components/header'

function App() {
  const { errStore } = useStores()
  console.log(errStore)
  return useObserver(()=>
    <div className="app">
      <Header />
      <Errors />
    </div>
  )
}

// function Header(){
//   return useObserver(()=>
//     <header className="header">
//       Sphinx Memes
//     </header>
//   )
// }
function Errors(){
  const { errStore } = useStores()
  return useObserver(()=>
    <div className="errors">
      {errStore.errs.map(err=>{
        return <div className="error">
          <div>
            <span>{err.message}</span>
            <span>{dateformat(err.time, "dddd, mmmm dS, yyyy, h:MM:ss TT")}</span>
          </div>
          <div>
            <span>{err.details}</span>&nbsp;&nbsp;
          </div>
        </div>
      })}
    </div>
  )
}

export default App
