import { Route, Routes } from "react-router-dom";
import LoginPage from "src/features/login-page/LoginPage";
import LogoutPage from "src/features/logout-page/LogoutPage";
import MainPage from "src/features/main-page/MainPage";
import Navbar from "src/features/navbar/Navbar";
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
    </Routes>
  </>;
};

export default App;
