deployment
load bslocations.dat
figure
locs=complex(bslocations(:,2),bslocations(:,3));
locs=locs(1:length(locs)/3)
drawPolyGon(locs,3200,'r')
ID=VirtualCellID;
% ID=[0 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 0 13 12 16 15 14 18 17 16 8 7 18 10 9 8 12 11 10 15 4 3 11 17 5 4 13 7 6 5 15 9 1 6 17 11 2 1 7 13 3 2 9];
for k=1:length(locs)
 [bslocations(k,1) ID(k)]
    text(bslocations(k,2),bslocations(k,3),sprintf('%d,V%d',ID(k),k-1));
end
drawPolyGon(locs(1:19),3200,'g')