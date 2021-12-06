#include <iostream>
#include <vector>
#include <map>
#include <set>
#include <stack>
#include <algorithm>
#include <array>
#include <unordered_map>
#include <queue>
#include <unordered_set>
#include <iomanip>
#include <zconf.h>

#define pb push_back
#define sz(v) ((int)(v).size())
#define all(v) (v).begin(),(v).end()

using namespace std;

typedef long long int64;
typedef vector<int> vi;
typedef pair<int, int> ii;

//Given a string s, return the longest palindromic substring in s. 
//
// 
// Example 1: 
//
// 
//Input: s = "babad"
//Output: "bab"
//Note: "aba" is also a valid answer.
// 
//
// Example 2: 
//
// 
//Input: s = "cbbd"
//Output: "bb"
// 
//
// Example 3: 
//
// 
//Input: s = "a"
//Output: "a"
// 
//
// Example 4: 
//
// 
//Input: s = "ac"
//Output: "a"
// 
//
// 
// Constraints: 
//
// 
// 1 <= s.length <= 1000 
// s consist of only digits and English letters. 
// 
// Related Topics String Dynamic Programming 👍 13983 👎 825

class Solution {
public:
  string longestPalindrome(string s) {
    string ans;

    for (int i = 0; i < s.size(); i++) {
      int len1 = expandAroundCenter(s, i, i);
      int len2 = expandAroundCenter(s, i, i + 1);
      int mx = max(len1, len2);
      if (mx > ans.size()) {
        ans = s.substr(i - (mx-1)/2, mx);
      }
    }
    return ans;
  }

  int expandAroundCenter(string &s, int l, int r) {
      if (r >= s.size()) {
        return 0;
      }
      while (l >= 0 && r < s.size() && s[l] == s[r]) {
        l--, r++;
      }
      return r - l - 1;
    }
};

int main() {
  ios::sync_with_stdio(false);
  cin.tie(0);
  
  int t;
  cin >> t;
  while (t--) {
    auto sol = new Solution();
    string s;
    cin >> s;
    cout << sol->longestPalindrome(s) << endl;
  }
}
