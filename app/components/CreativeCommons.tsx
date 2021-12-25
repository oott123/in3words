import React from 'react'
import ccBySaIcon from '../assets/cc-by-sa.svg'
import ccByNcSaIcon from '../assets/cc-by-nc-sa.svg'
import { Link } from 'remix'

const CreativeCommons: React.FC = () => {
  return (
    <div className="CreativeCommons">
      <p className="CreativeCommons_Icon">
        <a
          href="https://creativecommons.org/licenses/by-sa/4.0/"
          target="_blank"
          rel="noopener noreferrer"
        >
          <img src={ccBySaIcon} alt="CC-BY-SA 4.0" />
        </a>
        <a
          href="https://creativecommons.org/licenses/by-nc-sa/4.0/"
          target="_blank"
          rel="noopener noreferrer"
        >
          <img src={ccByNcSaIcon} alt="CC-BY-SA 4.0" />
        </a>
      </p>
      <p>
        如非特别说明，本站所有文章均为原创，并以 CC BY-SA 或 BY-NC-SA 4.0
        授权你使用，具体请参考本站的
        <Link to="/license">内容使用许可协议</Link>。
      </p>
      <p>
        <small>
          好耶，本站内容是基于自由文化许可协议共享的！什么是
          <a
            href="https://creativecommons.org/share-your-work/public-domain/freeworks"
            target="_blank"
            rel="noopener noreferrer"
          >
            自由文化许可协议
          </a>
          ？
        </small>
      </p>
    </div>
  )
}

export default CreativeCommons
