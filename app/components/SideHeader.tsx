import React from 'react'

const SideHeader: React.FC = ({ children }) => {
  return (
    <div className="SideHeader">
      <h3>{children}</h3>
    </div>
  )
}

export default SideHeader
