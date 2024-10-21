import Link from "next/link"
import "./Navbar.css"
import styles from "../page.module.css";

export default function Navbar() {  
  return (
    <div style={{backgroundColor: "black"}}>
      <div id="Navbar">
        <h1>GameDir</h1>
        <a>Top games</a>
        <a>Top reviews</a>

        <div className="right">
          <button className={styles.button}>Login</button>
          <button>Signup</button>
        </div>
      </div>
    </div>
  )
}