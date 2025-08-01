import { Route, Routes } from "react-router-dom";
import ContactsPage from "src/features/contacts-page/ContactsPage";
import LoadContactsPage from "src/features/load-contacts-page/LoadContactsPage";
import LoginPage from "src/features/login-page/LoginPage";
import LogoutPage from "src/features/logout-page/LogoutPage";
import MainPage from "src/features/main-page/MainPage";
import Navbar from "src/features/navbar/Navbar";
import SendConfirmPage from "src/features/send-page/SendConfirmPage";
import SendPage from "src/features/send-page/SendPage";
import SignupPage from "src/features/signup-page/SignupPage";
import TemplatePage from "src/features/template-page/TemplatePage";
import "./styles.css";

const App = () => {
  return <>
    <Navbar/>

    <Routes>
      <Route path="/" element={<MainPage/>}/>
      <Route path="/signup" element={<SignupPage/>}/>
      <Route path="/login" element={<LoginPage/>}/>
      <Route path="/logout" element={<LogoutPage/>}/>
      <Route path="/template" element={<TemplatePage/>}/>
      <Route path="/contacts" element={<ContactsPage/>}/>
      <Route path="/load-contacts" element={<LoadContactsPage/>}/>
      <Route path="/send" element={<SendPage/>}/>
      <Route path="/send/:id" element={<SendConfirmPage/>}/>
    </Routes>
  </>;
};

export default App;
