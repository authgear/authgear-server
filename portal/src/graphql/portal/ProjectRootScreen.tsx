import React, { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useSystemConfig } from "../../context/SystemConfigContext";
import ShowLoading from "../../ShowLoading";

const ProjectRootScreen: React.FC = function ProjectRootScreen() {
  const { analyticEnabled } = useSystemConfig();
  const navigate = useNavigate();

  useEffect(() => {
    navigate(analyticEnabled ? "./analytics" : "./users/", { replace: true });
  }, [navigate, analyticEnabled]);

  return <ShowLoading />;
};

export default ProjectRootScreen;
