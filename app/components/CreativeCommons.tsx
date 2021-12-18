import React from 'react'
import ccBySaIcon from '../assets/cc-by-sa.svg'

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
      </p>
      <p>
        如非特别说明，本站所有文章均为原创，并遵循{' '}
        <a
          href="https://creativecommons.org/licenses/by-sa/4.0/"
          target="_blank"
          rel="noopener noreferrer"
        >
          CC-BY-SA（署名-相同方式共享）4.0
        </a>{' '}
        协议发布。
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
