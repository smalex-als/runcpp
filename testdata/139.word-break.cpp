class Solution {
public:
    bool wordBreak(string s, vector<string>& wordDict) {
      vector<bool> dp = vector<bool>(s.size());
      dp[0] = true;
      for (int i = 0; i < s.size(); i++) {
        if (dp[i]) {
          for (string w : wordDict) {
            if (s.substr(i, w.size()) == w) {
              dp[i + w.size()] = true;
            }
          }
        }
      }
      return dp[s.size()];
    }
};
