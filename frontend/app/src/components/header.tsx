import React, {useState} from 'react';
import { useObserver } from 'mobx-react-lite'
import { useStores } from '../store'
import { orderBy } from 'lodash';

import {
  EuiHeader,
  EuiHeaderBreadcrumbs,
  EuiPopover,
  EuiPopoverTitle,
  EuiSelectable,
  EuiHeaderSection,
  EuiHeaderSectionItem,
  EuiHeaderSectionItemButton,
  EuiHeaderLogo,
  EuiButton,
  EuiIcon,
  EuiFieldSearch,
  EuiComboBox,
  EuiComboBoxOptionProps,
} from '@elastic/eui';

// import HeaderAppMenu from './header_app_menu';
// import HeaderUserMenu from './header_user_menu';
// import HeaderSpacesMenu from './header_spaces_menu';

export default function Header() {
  const { errStore } = useStores()
  const [text, setText] = useState<string>('')
  const [selectedTags, setSelectedTags] = useState<any[]>([
    {label:'hi'},{label:'lo'}
  ])
  const [tagsPop, setTagsPop] = useState(false)

  const button = (
    <EuiButton
      iconType="arrowDown"
      iconSide="right"
      size="s"
      onClick={()=>{
        setSelectedTags(orderBy(selectedTags, ['checked'], ['asc']))
        setTagsPop(!tagsPop)
      }}>
      Tags
    </EuiButton>
  );

  return useObserver(() =>
    <EuiHeader style={{justifyContent:'space-between',alignItems:'center',maxHeight:50,height:50,minHeight:50}}>
      <EuiHeaderSection grow={false}>
        <EuiHeaderSectionItem border="right">
          hi
        </EuiHeaderSectionItem>
        <EuiHeaderSectionItem border="right">
          {/* <HeaderSpacesMenu /> */}
        </EuiHeaderSectionItem>
      </EuiHeaderSection>

      <EuiHeaderSection side="right" style={{display:'flex',alignItems:'center'}}>
        {/* <EuiHeaderSectionItem> */}
        <EuiPopover
          panelPaddingSize="none"
          button={button}
          isOpen={tagsPop}
          closePopover={()=>setTagsPop(false)}>
          <EuiSelectable
            options={selectedTags}
            onChange={opts=>{
              console.log(opts)
              setSelectedTags(opts)
            }}>
            {(list, search) => (
              <div style={{ width: 240 }}>
                {list}
              </div>
            )}
          </EuiSelectable>
        </EuiPopover>
        <div style={{margin:'0 6px'}}>
          <EuiFieldSearch
            style={{width:'50vw'}}
            placeholder="Search for Memes"
            value={text}
            onChange={e=> setText(e.target.value)}
            // isClearable={this.state.isClearable}
            aria-label="Use aria labels when no actual label is in use"
          />
        </div>
      </EuiHeaderSection>

    </EuiHeader>
  )
}
