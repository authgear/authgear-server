"use strict";

Object.defineProperty(exports, "__esModule", {
  value: true
});
exports.Config = void 0;

var _logger = require("@parcel/logger");

var _path = _interopRequireDefault(require("path"));

var _fs = _interopRequireDefault(require("fs"));

function _interopRequireDefault(obj) { return obj && obj.__esModule ? obj : { default: obj }; }

function _defineProperty(obj, key, value) { if (key in obj) { Object.defineProperty(obj, key, { value: value, enumerable: true, configurable: true, writable: true }); } else { obj[key] = value; } return obj; }

const PACKAGE_JSON_SECTION = "parcel-namer-rewrite";

class Config {
  /**
   * Disable namer in development
   */

  /**
   * Disable name hashing in development
   */

  /**
   * Disable logging names
   */

  /**
   * Use file name hashes from parcel
   */
  constructor() {
    _defineProperty(this, "rules", void 0);

    _defineProperty(this, "chain", void 0);

    _defineProperty(this, "developmentDisable", false);

    _defineProperty(this, "developmentHashing", false);

    _defineProperty(this, "silent", false);

    _defineProperty(this, "useParcelHash", true);

    this.chain = '@parcel/namer-default';
    this.rules = [];
  }

  loadFromPackageFolder(rootFolder, logger) {
    const packageJson = _fs.default.readFileSync(_path.default.join(rootFolder, 'package.json')).toString();

    const packageInfo = JSON.parse(packageJson);
    const packageSection = packageInfo[PACKAGE_JSON_SECTION];

    if (!packageSection) {
      logger.warn({
        message: `no "${PACKAGE_JSON_SECTION}" section in package.json. Use no-rules config`
      });
      return;
    }

    if (packageSection && 'chain' in packageSection) {
      this.chain = packageSection.chain;
    }

    this.silent = packageSection && 'silent' in packageSection && packageSection.silent;

    if (packageSection && 'useParcelHash' in packageSection) {
      this.useParcelHash = !!packageSection.useParcelHash;
    }

    if (packageSection && 'rules' in packageSection) {
      Object.keys(packageSection.rules).forEach(k => {
        const ruleData = packageSection.rules[k];
        const ruleTo = typeof ruleData === 'string' ? ruleData : null;

        if (ruleTo === null) {
          logger.warn(`No "to" rule for test "${k}" `);
          return;
        }

        this.rules.push({
          test: new RegExp(k),
          to: ruleTo
        });
      });
    }

    if (packageSection && 'developmentHashing' in packageSection) {
      this.developmentHashing = !!packageSection.developmentHashing;
    }

    if (packageSection && 'developmentDisable' in packageSection) {
      this.developmentDisable = !!packageSection.developmentDisable;
    }
  }

  selectRule(name) {
    const matches = this.rules.map(rule => rule.test.test(name) ? rule : null).filter(rule => rule != null);

    if (matches.length > 0) {
      return matches[0];
    }

    return null;
  }

}

exports.Config = Config;
//# sourceMappingURL=data:application/json;charset=utf-8;base64,eyJ2ZXJzaW9uIjozLCJuYW1lcyI6WyJQQUNLQUdFX0pTT05fU0VDVElPTiIsIkNvbmZpZyIsImNvbnN0cnVjdG9yIiwiY2hhaW4iLCJydWxlcyIsImxvYWRGcm9tUGFja2FnZUZvbGRlciIsInJvb3RGb2xkZXIiLCJsb2dnZXIiLCJwYWNrYWdlSnNvbiIsImZzIiwicmVhZEZpbGVTeW5jIiwicGF0aCIsImpvaW4iLCJ0b1N0cmluZyIsInBhY2thZ2VJbmZvIiwiSlNPTiIsInBhcnNlIiwicGFja2FnZVNlY3Rpb24iLCJ3YXJuIiwibWVzc2FnZSIsInNpbGVudCIsInVzZVBhcmNlbEhhc2giLCJPYmplY3QiLCJrZXlzIiwiZm9yRWFjaCIsImsiLCJydWxlRGF0YSIsInJ1bGVUbyIsInB1c2giLCJ0ZXN0IiwiUmVnRXhwIiwidG8iLCJkZXZlbG9wbWVudEhhc2hpbmciLCJkZXZlbG9wbWVudERpc2FibGUiLCJzZWxlY3RSdWxlIiwibmFtZSIsIm1hdGNoZXMiLCJtYXAiLCJydWxlIiwiZmlsdGVyIiwibGVuZ3RoIl0sInNvdXJjZXMiOlsiLi4vc3JjL0NvbmZpZy5qcyJdLCJzb3VyY2VzQ29udGVudCI6WyJpbXBvcnQge1BsdWdpbkxvZ2dlcn0gZnJvbSAnQHBhcmNlbC9sb2dnZXInO1xuaW1wb3J0IHBhdGggZnJvbSAncGF0aCc7XG5pbXBvcnQgZnMgZnJvbSAnZnMnO1xuXG5jb25zdCBQQUNLQUdFX0pTT05fU0VDVElPTiA9IFwicGFyY2VsLW5hbWVyLXJld3JpdGVcIjtcblxuZXhwb3J0IGNsYXNzIENvbmZpZyB7XG4gICAgcnVsZXM6IE5hbWVyUnVsZVtdXG4gICAgY2hhaW46IHN0cmluZ1xuICAgIC8qKlxuICAgICAqIERpc2FibGUgbmFtZXIgaW4gZGV2ZWxvcG1lbnRcbiAgICAgKi9cbiAgICBkZXZlbG9wbWVudERpc2FibGUgPSBmYWxzZVxuICAgIC8qKlxuICAgICAqIERpc2FibGUgbmFtZSBoYXNoaW5nIGluIGRldmVsb3BtZW50XG4gICAgICovXG4gICAgZGV2ZWxvcG1lbnRIYXNoaW5nID0gZmFsc2VcbiAgICAvKipcbiAgICAgKiBEaXNhYmxlIGxvZ2dpbmcgbmFtZXNcbiAgICAgKi9cbiAgICBzaWxlbnQgPSBmYWxzZVxuICAgIC8qKlxuICAgICAqIFVzZSBmaWxlIG5hbWUgaGFzaGVzIGZyb20gcGFyY2VsXG4gICAgICovXG4gICAgdXNlUGFyY2VsSGFzaCA9IHRydWVcblxuICAgIGNvbnN0cnVjdG9yKCkge1xuICAgICAgICB0aGlzLmNoYWluID0gJ0BwYXJjZWwvbmFtZXItZGVmYXVsdCc7XG4gICAgICAgIHRoaXMucnVsZXMgPSBbXTtcbiAgICB9XG5cbiAgICBsb2FkRnJvbVBhY2thZ2VGb2xkZXIocm9vdEZvbGRlcjogc3RyaW5nLCBsb2dnZXI6IFBsdWdpbkxvZ2dlcikge1xuICAgICAgICBjb25zdCBwYWNrYWdlSnNvbiA9IGZzLnJlYWRGaWxlU3luYyhwYXRoLmpvaW4ocm9vdEZvbGRlciwgJ3BhY2thZ2UuanNvbicpKS50b1N0cmluZygpO1xuICAgICAgICBjb25zdCBwYWNrYWdlSW5mbyA9IEpTT04ucGFyc2UocGFja2FnZUpzb24pO1xuICAgICAgICBjb25zdCBwYWNrYWdlU2VjdGlvbiA9IHBhY2thZ2VJbmZvW1BBQ0tBR0VfSlNPTl9TRUNUSU9OXTtcbiAgICAgICAgaWYgKCFwYWNrYWdlU2VjdGlvbikge1xuICAgICAgICAgICAgbG9nZ2VyLndhcm4oe1xuICAgICAgICAgICAgICAgIG1lc3NhZ2U6IGBubyBcIiR7UEFDS0FHRV9KU09OX1NFQ1RJT059XCIgc2VjdGlvbiBpbiBwYWNrYWdlLmpzb24uIFVzZSBuby1ydWxlcyBjb25maWdgXG4gICAgICAgICAgICB9KVxuICAgICAgICAgICAgcmV0dXJuO1xuICAgICAgICB9XG5cbiAgICAgICAgaWYgKHBhY2thZ2VTZWN0aW9uICYmICdjaGFpbicgaW4gcGFja2FnZVNlY3Rpb24pIHtcbiAgICAgICAgICAgIHRoaXMuY2hhaW4gPSBwYWNrYWdlU2VjdGlvbi5jaGFpbjtcbiAgICAgICAgfVxuXG4gICAgICAgIHRoaXMuc2lsZW50ID0gcGFja2FnZVNlY3Rpb24gJiYgJ3NpbGVudCcgaW4gcGFja2FnZVNlY3Rpb24gJiYgcGFja2FnZVNlY3Rpb24uc2lsZW50O1xuXG4gICAgICAgIGlmIChwYWNrYWdlU2VjdGlvbiAmJiAndXNlUGFyY2VsSGFzaCcgaW4gcGFja2FnZVNlY3Rpb24pIHtcbiAgICAgICAgICAgIHRoaXMudXNlUGFyY2VsSGFzaCA9ICEhcGFja2FnZVNlY3Rpb24udXNlUGFyY2VsSGFzaDtcbiAgICAgICAgfVxuXG4gICAgICAgIGlmIChwYWNrYWdlU2VjdGlvbiAmJiAncnVsZXMnIGluIHBhY2thZ2VTZWN0aW9uKSB7XG4gICAgICAgICAgICBPYmplY3Qua2V5cyhwYWNrYWdlU2VjdGlvbi5ydWxlcykuZm9yRWFjaChrID0+IHtcbiAgICAgICAgICAgICAgICBjb25zdCBydWxlRGF0YSA9IHBhY2thZ2VTZWN0aW9uLnJ1bGVzW2tdO1xuICAgICAgICAgICAgICAgIGNvbnN0IHJ1bGVUbyA9IHR5cGVvZiBydWxlRGF0YSA9PT0gJ3N0cmluZycgPyBydWxlRGF0YSA6IG51bGw7XG4gICAgICAgICAgICAgICAgaWYgKHJ1bGVUbyA9PT0gbnVsbCkge1xuICAgICAgICAgICAgICAgICAgICBsb2dnZXIud2FybihgTm8gXCJ0b1wiIHJ1bGUgZm9yIHRlc3QgXCIke2t9XCIgYCk7XG4gICAgICAgICAgICAgICAgICAgIHJldHVybjtcbiAgICAgICAgICAgICAgICB9XG5cbiAgICAgICAgICAgICAgICB0aGlzLnJ1bGVzLnB1c2goe1xuICAgICAgICAgICAgICAgICAgICB0ZXN0OiBuZXcgUmVnRXhwKGspLFxuICAgICAgICAgICAgICAgICAgICB0bzogcnVsZVRvXG4gICAgICAgICAgICAgICAgfSlcbiAgICAgICAgICAgIH0pXG4gICAgICAgIH1cblxuICAgICAgICBpZiAocGFja2FnZVNlY3Rpb24gJiYgJ2RldmVsb3BtZW50SGFzaGluZycgaW4gcGFja2FnZVNlY3Rpb24pIHtcbiAgICAgICAgICAgIHRoaXMuZGV2ZWxvcG1lbnRIYXNoaW5nID0gISFwYWNrYWdlU2VjdGlvbi5kZXZlbG9wbWVudEhhc2hpbmc7XG4gICAgICAgIH1cblxuICAgICAgICBpZiAocGFja2FnZVNlY3Rpb24gJiYgJ2RldmVsb3BtZW50RGlzYWJsZScgaW4gcGFja2FnZVNlY3Rpb24pIHtcbiAgICAgICAgICAgIHRoaXMuZGV2ZWxvcG1lbnREaXNhYmxlID0gISFwYWNrYWdlU2VjdGlvbi5kZXZlbG9wbWVudERpc2FibGU7XG4gICAgICAgIH1cbiAgICB9XG5cbiAgICBzZWxlY3RSdWxlKG5hbWU6IHN0cmluZyk6IE5hbWVyUnVsZSB8IG51bGwge1xuICAgICAgICBjb25zdCBtYXRjaGVzID0gdGhpcy5ydWxlc1xuICAgICAgICAgICAgLm1hcChydWxlID0+IHJ1bGUudGVzdC50ZXN0KG5hbWUpID8gcnVsZSA6IG51bGwpXG4gICAgICAgICAgICAuZmlsdGVyKHJ1bGUgPT4gcnVsZSAhPSBudWxsKTtcbiAgICAgICAgaWYgKG1hdGNoZXMubGVuZ3RoID4gMCkge1xuICAgICAgICAgICAgcmV0dXJuIG1hdGNoZXNbMF07XG4gICAgICAgIH1cbiAgICAgICAgcmV0dXJuIG51bGw7XG4gICAgfVxufVxuXG5leHBvcnQgaW50ZXJmYWNlIE5hbWVyUnVsZSB7XG4gICAgdGVzdDogUmVnRXhwO1xuICAgIHRvOiBzdHJpbmc7XG59XG4iXSwibWFwcGluZ3MiOiI7Ozs7Ozs7QUFBQTs7QUFDQTs7QUFDQTs7Ozs7O0FBRUEsTUFBTUEsb0JBQW9CLEdBQUcsc0JBQTdCOztBQUVPLE1BQU1DLE1BQU4sQ0FBYTtFQUdoQjtBQUNKO0FBQ0E7O0VBRUk7QUFDSjtBQUNBOztFQUVJO0FBQ0o7QUFDQTs7RUFFSTtBQUNKO0FBQ0E7RUFHSUMsV0FBVyxHQUFHO0lBQUE7O0lBQUE7O0lBQUEsNENBZE8sS0FjUDs7SUFBQSw0Q0FWTyxLQVVQOztJQUFBLGdDQU5MLEtBTUs7O0lBQUEsdUNBRkUsSUFFRjs7SUFDVixLQUFLQyxLQUFMLEdBQWEsdUJBQWI7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