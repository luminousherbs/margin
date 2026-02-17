import React from "react";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { initAuth, $user } from "./store/auth";
import { loadPreferences } from "./store/preferences";
import { useStore } from "@nanostores/react";

import AppLayout from "./layouts/AppLayout";
import Feed from "./views/core/Feed";
import Login from "./views/auth/Login";
import Notifications from "./views/core/Notifications";
import Collections from "./views/collections/Collections";
import Settings from "./views/core/Settings";
import NewAnnotationPage from "./views/core/New";
import MasonryFeed from "./components/feed/MasonryFeed";
import {
  ProfileWrapper,
  SelfProfileWrapper,
  CollectionDetailWrapper,
  AnnotationDetailWrapper,
  UserUrlWrapper,
  UrlWrapper,
} from "./routes/wrappers";
import About from "./views/About";
import AdminModeration from "./views/core/AdminModeration";

function RootRoute() {
  const user = useStore($user);

  if (user) {
    return <Navigate to="/home" replace />;
  }

  return <About />;
}

export default function App() {
  React.useEffect(() => {
    initAuth();
    loadPreferences();
  }, []);

  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<RootRoute />} />
        <Route path="/login" element={<Login />} />
        <Route path="/about" element={<About />} />
        <Route path="/auth/*" element={<div>Redirecting...</div>} />

        <Route
          path="/home"
          element={
            <AppLayout>
              <Feed initialType="all" />
            </AppLayout>
          }
        />
        <Route path="/my-feed" element={<Navigate to="/home" replace />} />

        <Route
          path="/annotations"
          element={
            <AppLayout>
              <MasonryFeed
                motivation="commenting"
                emptyMessage="You haven't annotated anything yet."
                showTabs={true}
                title="Annotations"
              />
            </AppLayout>
          }
        />
        <Route
          path="/bookmarks"
          element={
            <AppLayout>
              <MasonryFeed
                motivation="bookmarking"
                emptyMessage="You haven't bookmarked anything yet."
                showTabs={true}
                title="Bookmarks"
              />
            </AppLayout>
          }
        />
        <Route
          path="/highlights"
          element={
            <AppLayout>
              <MasonryFeed
                motivation="highlighting"
                emptyMessage="You haven't highlighted anything yet."
                showTabs={true}
                title="Highlights"
              />
            </AppLayout>
          }
        />

        <Route
          path="/collections"
          element={
            <AppLayout>
              <Collections />
            </AppLayout>
          }
        />
        <Route
          path="/:handle/collection/:rkey"
          element={
            <AppLayout>
              <CollectionDetailWrapper />
            </AppLayout>
          }
        />
        <Route
          path="/collections/:rkey"
          element={
            <AppLayout>
              <CollectionDetailWrapper />
            </AppLayout>
          }
        />

        <Route
          path="/profile/:did"
          element={
            <AppLayout>
              <ProfileWrapper />
            </AppLayout>
          }
        />
        <Route
          path="/profile"
          element={
            <AppLayout>
              <SelfProfileWrapper />
            </AppLayout>
          }
        />

        <Route
          path="/new"
          element={
            <AppLayout>
              <NewAnnotationPage />
            </AppLayout>
          }
        />
        <Route
          path="/at/:did/:rkey"
          element={
            <AppLayout>
              <AnnotationDetailWrapper />
            </AppLayout>
          }
        />
        <Route
          path="/annotation/:uri"
          element={
            <AppLayout>
              <AnnotationDetailWrapper />
            </AppLayout>
          }
        />
        <Route
          path="/:handle/annotation/:rkey"
          element={
            <AppLayout>
              <AnnotationDetailWrapper />
            </AppLayout>
          }
        />
        <Route
          path="/:handle/highlight/:rkey"
          element={
            <AppLayout>
              <AnnotationDetailWrapper />
            </AppLayout>
          }
        />
        <Route
          path="/:handle/bookmark/:rkey"
          element={
            <AppLayout>
              <AnnotationDetailWrapper />
            </AppLayout>
          }
        />
        <Route
          path="/:handle/url/*"
          element={
            <AppLayout>
              <UserUrlWrapper />
            </AppLayout>
          }
        />
        <Route
          path="/url/*"
          element={
            <AppLayout>
              <UrlWrapper />
            </AppLayout>
          }
        />

        <Route
          path="/admin/moderation"
          element={
            <AppLayout>
              <AdminModeration />
            </AppLayout>
          }
        />

        <Route
          path="/settings"
          element={
            <AppLayout>
              <Settings />
            </AppLayout>
          }
        />
        <Route
          path="/notifications"
          element={
            <AppLayout>
              <Notifications />
            </AppLayout>
          }
        />

        <Route path="*" element={<Navigate to="/home" replace />} />
      </Routes>
    </BrowserRouter>
  );
}
