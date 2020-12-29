{-# OPTIONS_GHC -Wall #-}
{-# LANGUAGE OverloadedStrings #-}
module Socket (watchFile) where

import Control.Concurrent (forkIO, threadDelay)
import qualified Data.ByteString.Char8 as BS
import qualified Network.WebSockets as WS
import qualified System.FSNotify.Devel as Notify


watchFile :: FilePath -> WS.PendingConnection -> IO ()
watchFile watchedFile pendingConnection =
  do  connection <- WS.acceptRequest pendingConnection

      Notify.withManager $ \mgmt ->
        do  stop <- Notify.treeExtAny mgmt "." ".elm" print
            tend connection
            stop
