import Image from "next/image";
import styles from "./page.module.css";

function FriendsActivity() {
  return (
    <>
    </>
  )
}

function RecentReleases() {
  return (
    <>
      <h1 className={styles.page}>Watda</h1>
    </>
  )
}

export default function Home() {
  return (
    <div className={styles.page}>
      <h1>Hello world</h1>
      <RecentReleases />
    </div>
  );
}
